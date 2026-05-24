package episodex

import (
	"regexp"
	"strconv"
	"strings"
)

type indexedRegexp struct {
	group int
	re    *regexp.Regexp
}

// EpisodeNumber 描述可用于比较和偏移的剧集编号。
type EpisodeNumber struct {
	Raw   string
	Major int
	Minor int
}

// Offset 返回偏移 delta 集后的剧集编号字符串，保留小数部分。
func (n EpisodeNumber) Offset(delta int) string {
	return formatEpisodeNumber(n.Major+delta, n.Minor)
}

// EpisodeInfo 描述文件名中的季、单集、集数范围和全集信息。
//
// 对单集资源，Episode 会有值；对合集/整季资源，Start/End 表示范围，
// Complete 表示命中了全集/全 N 集/Fin/END 等整季语义。
type EpisodeInfo struct {
	Season   *int
	Episode  *EpisodeNumber
	Start    *EpisodeNumber
	End      *EpisodeNumber
	Complete bool
}

var (
	bracketPairs = [][2]string{
		{`\[`, `\]`},
		{`\(`, `\)`},
		{`（`, `）`},
		{`【`, `】`},
		{`「`, `」`},
		{`〔`, `〕`},
	}

	combinedEpisodePatterns = []struct {
		group   int
		pattern string
	}{
		{1, `(\d{1,4}(?:\.5)?)(?:[Vv]\d{1,2})?\s*_?\s*(?i:END)?`},
		{1, `第\s*(\d{1,4}(?:\.5)?)\s*[集话話]`},
		{2, `([Ee][Pp]|[Ss][Pp]|[Ee])\s*[-_.]?\s*(\d{1,4}(?:\.5)?)(?:[Vv]\d{1,2})?\s*_?\s*(?i:END)?`},
	}

	combinedEpisodeRE []indexedRegexp

	cleanBracketRe = regexp.MustCompile(`\[[^\]]*\]|\([^\)]*\)|（[^）]*）|【[^】]*】|「[^」]*」|〔[^〕]*〕`)
	cleanEpisodeRE = []indexedRegexp{
		{group: 1, re: regexp.MustCompile(`第\s*(\d{1,4}(?:\.5)?)\s*[集话話]`)},
		{group: 2, re: regexp.MustCompile(`([Ee][Pp]|[Ss][Pp]|[Ee])\s*[-_.]?\s*(\d{1,4}(?:\.5)?)`)},
	}

	cleanTrashRe  = regexp.MustCompile(`(?i)\b(?:1080p|720p|2160p|4k|2k|8k|x264|x265|h264|h265|10bit|8bit|mp4|m4v|m2ts|3gp)\b`)
	cleanSeasonRe = regexp.MustCompile(`(?i)\bS\s*\d{1,3}\b|第\s*\d{1,3}\s*季|第\s*[零〇一二三四五六七八九十两]{1,3}\s*季`)
	cleanYearRe   = regexp.MustCompile(`\b(?:19|20)\d{2}\b`)
	reV           = regexp.MustCompile(`(\d+(?:\.5)?)[Vv]\d+`)
	reDecimal     = regexp.MustCompile(`\b\d{1,3}\.5\b`)
	reLastDigits  = regexp.MustCompile(`\d+`)

	seasonDirectRe  = regexp.MustCompile(`(?i)\bS\s*(\d{1,3})\b`)
	seasonCnRe      = regexp.MustCompile(`第\s*(\d{1,3})\s*季`)
	seasonCnNumRe   = regexp.MustCompile(`第\s*([零〇一二三四五六七八九十两]{1,3})\s*季`)
	seasonEpisodeRe = regexp.MustCompile(`(?i)\bS\s*(\d{1,3})\s*[-_. ]*E\s*(\d{1,4}(?:\.5)?)\b`)

	episodeRangeRe       = regexp.MustCompile(`(?i)(?:^|[^\d])(\d{1,4}(?:\.5)?)\s*(?:-|~|–|—|－|至|到)\s*(\d{1,4}(?:\.5)?)(?:\s*(?:fin|end))?(?:[^\d]|$)`)
	fullCountRe          = regexp.MustCompile(`全\s*(\d{1,4}|[零〇一二三四五六七八九十两百]{1,5})\s*[集话話]`)
	completeSeasonMarkRe = regexp.MustCompile(`(?i)全集|全季|完结|完結|fin|end`)
)

func init() {
	for _, item := range combinedEpisodePatterns {
		for _, pair := range bracketPairs {
			p := pair[0] + item.pattern + pair[1]
			combinedEpisodeRE = append(combinedEpisodeRE, indexedRegexp{
				group: item.group,
				re:    regexp.MustCompile(p),
			})
		}
	}
}

// ExtractEpisodeInfo 从发布名/文件名中提取季、单集、集数范围和全集语义。
//
// 示例：
//   - "SPY x FAMILY (01-12Fin WEBRIP...)" => Start=1 End=12 Complete=true
//   - "全职猎人2011[全62集]..." => Start=1 End=62 Complete=true
//   - "Show S01E02.mkv" => Season=1 Episode=2
func ExtractEpisodeInfo(name string) EpisodeInfo {
	if name == "" {
		return EpisodeInfo{}
	}

	info := EpisodeInfo{Season: parseIntPtr(extractSeason(name))}

	if start, end := extractEpisodeRange(name); start != "" || end != "" {
		info.Start = parseEpisodeNumberPtr(start)
		info.End = parseEpisodeNumberPtr(end)
		info.Complete = hasCompleteSeasonMark(name)
		return info
	}

	if end := extractFullCount(name); end != "" {
		info.Start = parseEpisodeNumberPtr("1")
		info.End = parseEpisodeNumberPtr(end)
		info.Complete = true
		return info
	}

	if season, episode := extractBySeasonEpisode(name); season != "" || episode != "" {
		info.Season = parseIntPtr(season)
		info.Episode = parseEpisodeNumberPtr(episode)
	} else {
		info.Season = parseIntPtr(extractSeason(name))
		info.Episode = parseEpisodeNumberPtr(extractEpisode(name))
	}
	info.Complete = hasCompleteSeasonMark(name)
	return info
}

func extractBySeasonEpisode(stem string) (string, string) {
	m := seasonEpisodeRe.FindStringSubmatch(stem)
	if len(m) >= 3 {
		return m[1], m[2]
	}
	return "", ""
}

func extractSeason(stem string) string {
	if m := seasonCnRe.FindStringSubmatch(stem); len(m) >= 2 {
		return m[1]
	}
	if m := seasonCnNumRe.FindStringSubmatch(stem); len(m) >= 2 {
		if n, ok := chineseNumberToInt(m[1]); ok {
			return n
		}
	}
	if m := seasonDirectRe.FindStringSubmatch(stem); len(m) >= 2 {
		return m[1]
	}
	return ""
}

func extractEpisode(stem string) string {
	// 命中显式 SxxEyy 时，直接返回 E 对应的集号。
	if _, ep := extractBySeasonEpisode(stem); ep != "" {
		return ep
	}

	// 在括号内部匹配
	for _, item := range combinedEpisodeRE {
		if caps := item.re.FindStringSubmatch(stem); len(caps) > item.group {
			return caps[item.group]
		}
	}

	// 括号内未匹配到，删除括号内容
	cleanName := cleanBracketRe.ReplaceAllString(stem, "")
	// 匹配括号外内容
	for _, item := range cleanEpisodeRE {
		if caps := item.re.FindStringSubmatch(cleanName); len(caps) > item.group {
			return caps[item.group]
		}
	}

	// 开始匹配纯数字
	// 模式: 10v1, 10.5v1
	if caps := reV.FindStringSubmatch(cleanName); len(caps) >= 2 {
		return caps[1]
	}
	// 模式：纯xx.5这种格式(前后不能有英文,最多只匹配到xxx.5)
	if m := reDecimal.FindString(cleanName); m != "" {
		return m
	}

	// 排除干扰项
	cleanForDigits := cleanTrashRe.ReplaceAllString(cleanName, "")
	// 排除季标记，避免把 S02 / 第2季 里的季号当成集号。
	cleanForDigits = cleanSeasonRe.ReplaceAllString(cleanForDigits, " ")
	// 排除年份，避免把 2019/2020 误识别为集号。
	cleanForDigits = cleanYearRe.ReplaceAllString(cleanForDigits, " ")
	// 匹配最后的数字
	allDigits := reLastDigits.FindAllString(cleanForDigits, -1)
	if len(allDigits) > 0 {
		return allDigits[len(allDigits)-1]
	}
	return ""
}

func extractEpisodeRange(stem string) (string, string) {
	matches := episodeRangeRe.FindAllStringSubmatch(stem, -1)
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		start, end := m[1], m[2]
		if isLikelyYear(start) || isLikelyYear(end) {
			continue
		}
		if lessOrEqualEpisodeIndex(start, end) {
			return start, end
		}
	}
	return "", ""
}

func extractFullCount(stem string) string {
	if m := fullCountRe.FindStringSubmatch(stem); len(m) >= 2 {
		if isDigits(m[1]) {
			return m[1]
		}
		if n, ok := chineseCountToInt(m[1]); ok {
			return strconv.Itoa(n)
		}
	}
	return ""
}

func hasCompleteSeasonMark(stem string) bool {
	if completeSeasonMarkRe.MatchString(stem) {
		return true
	}
	return fullCountRe.MatchString(stem)
}

func isLikelyYear(s string) bool {
	if len(s) != 4 {
		return false
	}
	n, err := strconv.Atoi(s)
	return err == nil && n >= 1900 && n <= 2099
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func lessOrEqualEpisodeIndex(start, end string) bool {
	a, errA := strconv.ParseFloat(start, 64)
	b, errB := strconv.ParseFloat(end, 64)
	return errA == nil && errB == nil && a <= b
}

func parseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}

func parseEpisodeNumberPtr(s string) *EpisodeNumber {
	n, ok := parseEpisodeNumber(s)
	if !ok {
		return nil
	}
	return &n
}

func parseEpisodeNumber(s string) (EpisodeNumber, bool) {
	if s == "" {
		return EpisodeNumber{}, false
	}
	parts := strings.SplitN(s, ".", 2)
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return EpisodeNumber{}, false
	}
	minor := 0
	if len(parts) == 2 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return EpisodeNumber{}, false
		}
	}
	return EpisodeNumber{Raw: s, Major: major, Minor: minor}, true
}

func formatEpisodeNumber(major, minor int) string {
	raw := strconv.Itoa(major)
	if minor != 0 {
		raw += "." + strconv.Itoa(minor)
	}
	return raw
}

func chineseCountToInt(s string) (int, bool) {
	digitMap := map[rune]int{
		'零': 0, '〇': 0,
		'一': 1, '二': 2, '两': 2, '三': 3, '四': 4, '五': 5,
		'六': 6, '七': 7, '八': 8, '九': 9,
	}

	total := 0
	current := 0
	hasAny := false

	for _, r := range s {
		if d, ok := digitMap[r]; ok {
			current = d
			hasAny = true
			continue
		}
		switch r {
		case '百':
			hasAny = true
			if current == 0 {
				current = 1
			}
			total += current * 100
			current = 0
		case '十':
			hasAny = true
			if current == 0 {
				current = 1
			}
			total += current * 10
			current = 0
		default:
			return 0, false
		}
	}

	total += current
	return total, hasAny && total > 0 && total <= 999
}

// chineseNumberToInt 将常见中文数字（1-99，如 一/十/二十）转换为阿拉伯数字字符串。
func chineseNumberToInt(s string) (string, bool) {
	digitMap := map[rune]int{
		'零': 0, '〇': 0,
		'一': 1, '二': 2, '两': 2, '三': 3, '四': 4, '五': 5,
		'六': 6, '七': 7, '八': 8, '九': 9,
	}
	unitMap := map[rune]int{'十': 10}

	total := 0
	current := 0
	hasAny := false

	for _, r := range s {
		if d, ok := digitMap[r]; ok {
			current = d
			hasAny = true
			continue
		}
		if u, ok := unitMap[r]; ok {
			hasAny = true
			if current == 0 {
				current = 1
			}
			total += current * u
			current = 0
			continue
		}
		return "", false
	}

	if !hasAny {
		return "", false
	}
	total += current
	if total < 1 || total > 99 {
		return "", false
	}
	return strconv.Itoa(total), true
}

// normalizeIndex 将纯个位数字规范化为两位（如 3 -> 03）。
func normalizeIndex(v string) string {
	if len(v) == 1 && v[0] >= '0' && v[0] <= '9' {
		return "0" + v
	}
	return v
}
