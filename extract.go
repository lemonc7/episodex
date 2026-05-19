package episodex

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type indexedRegexp struct {
	group int
	re    *regexp.Regexp
}

var (
	bracketPairs = [][2]string{
		{`\[`, `\]`},
		{`\(`, `\)`},
		{`【`, `】`},
		{`「`, `」`},
	}

	combinedEpisodePatterns = []struct {
		group   int
		pattern string
	}{
		{1, `(\d{1,3}(?:\.5)?)(?:[Vv]\d{1})?\s?(?:_)?(?i:END)?`},
		{1, `第(\d{1,4}(?:\.5)?)[集话話]`},
		{2, `([Ee][Pp]|[Ss][Pp]|[Ee])(\d{1,4}(?:\.5)?)(?:[Vv]\d{1})?\s?(?:_)?(?i:END)?`},
	}

	combinedEpisodeRE []indexedRegexp

	cleanBracketRe = regexp.MustCompile(`\[[^\]]*\]|\([^\)]*\)|【[^】]*】|「[^」]*」`)
	cleanEpisodeRE = []indexedRegexp{
		{group: 1, re: regexp.MustCompile(`第(\d{1,4}(?:\.5)?)[集话話]`)},
		{group: 2, re: regexp.MustCompile(`([Ee][Pp]|[Ss][Pp]|[Ee])(\d{1,4}(?:\.5)?)`)},
	}

	cleanTrashRe  = regexp.MustCompile(`(?i)\b(?:1080p|720p|2160p|4k|2k|8k|x264|x265|h264|h265|10bit|8bit)\b`)
	cleanSeasonRe = regexp.MustCompile(`(?i)\bS\d{1,3}\b|第\d{1,3}季|第[零〇一二三四五六七八九十两]{1,3}季`)
	cleanYearRe   = regexp.MustCompile(`\b(?:19|20)\d{2}\b`)
	reV           = regexp.MustCompile(`(\d+(?:\.5)?)[Vv]\d+`)
	reDecimal     = regexp.MustCompile(`\b\d{1,3}\.5\b`)
	reLastDigits  = regexp.MustCompile(`\d+`)

	seasonDirectRe  = regexp.MustCompile(`(?i)\bS(\d{1,3})\b`)
	seasonCnRe      = regexp.MustCompile(`第(\d{1,3})季`)
	seasonCnNumRe   = regexp.MustCompile(`第([零〇一二三四五六七八九十两]{1,3})季`)
	seasonEpisodeRe = regexp.MustCompile(`(?i)\bS(\d{1,3})\s*[-_. ]*E(\d{1,4}(?:\.5)?)\b`)
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

// Extract 从文件名中提取季号和集号，返回 (season, episode)。
func Extract(name string) (season, episode string) {
	stem := fileStem(name)
	if stem == "" {
		return "", ""
	}

	// 优先处理标准 SxxEyy，避免后续规则干扰。
	if season, episode := extractBySeasonEpisode(stem); season != "" || episode != "" {
		return normalizeIndex(season), normalizeIndex(episode)
	}

	return normalizeIndex(extractSeason(stem)), normalizeIndex(extractEpisode(stem))
}

func ExtractSeason(name string) string {
	season, _ := Extract(name)
	return season
}

func ExtractEpisode(name string) string {
	_, episode := Extract(name)
	return episode
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

func fileStem(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	if stem == "" {
		return filename
	}
	return stem
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
