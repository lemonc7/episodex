package episodex

import (
	"reflect"
	"testing"
)

func TestExtractEpisodeInfo(t *testing.T) {
	tests := []struct {
		in       string
		wantInfo EpisodeInfo
	}{
		{
			in: "[悠哈璃羽字幕社&LoliHouse] 辉夜大小姐想让我告白 _ Kaguya-sama wa Kokurasetai - 全集 [WebRip 1080p HEVC-10bit AAC][简繁内封字幕]",
			wantInfo: EpisodeInfo{
				Complete: true,
			},
		},
		{
			in: "[BDrip] Kaguya-sama wa Kokurasetai S01 [7³ACG]",
			wantInfo: EpisodeInfo{
				Season: new(1),
			},
		},
		{
			in: "My.Show.S01E02.1080p.WEB-DL.mkv",
			wantInfo: EpisodeInfo{
				Season:  new(1),
				Episode: episodePtr("02", 2, 0),
			},
		},
		{
			in: "某动画 第2季 第10集.mp4",
			wantInfo: EpisodeInfo{
				Season:  new(2),
				Episode: episodePtr("10", 10, 0),
			},
		},
		{
			in: "番剧标题 [12 END][1080p].mkv",
			wantInfo: EpisodeInfo{
				Episode:  episodePtr("12", 12, 0),
				Complete: true,
			},
		},
		{
			in: "Anime - EP07 - x265.mkv",
			wantInfo: EpisodeInfo{
				Episode: episodePtr("07", 7, 0),
			},
		},
		{
			in: "Anime - EP 07 - x265.mkv",
			wantInfo: EpisodeInfo{
				Episode: episodePtr("07", 7, 0),
			},
		},
		{
			in: "Anime - E_07 - x265.mkv",
			wantInfo: EpisodeInfo{
				Episode: episodePtr("07", 7, 0),
			},
		},
		{
			in: "Anime.10v1.1080p.mkv",
			wantInfo: EpisodeInfo{
				Episode: episodePtr("10", 10, 0),
			},
		},
		{
			in: "Anime.10.5v1.BDRip.mkv",
			wantInfo: EpisodeInfo{
				Episode: episodePtr("10.5", 10, 5),
			},
		},
		{
			in: "S02",
			wantInfo: EpisodeInfo{
				Season: new(2),
			},
		},
		{
			in: "S 01 E 02.mkv",
			wantInfo: EpisodeInfo{
				Season:  new(1),
				Episode: episodePtr("02", 2, 0),
			},
		},
		{
			in: "第 12 话",
			wantInfo: EpisodeInfo{
				Episode: episodePtr("12", 12, 0),
			},
		},
		{
			in: "第十一季 100",
			wantInfo: EpisodeInfo{
				Season:  new(11),
				Episode: episodePtr("100", 100, 0),
			},
		},
		{
			in:       "2020",
			wantInfo: EpisodeInfo{},
		},
		{
			in:       "示例剧.第一百季.1080p.mkv",
			wantInfo: EpisodeInfo{},
		},
		{
			in:       "Movie.mp4",
			wantInfo: EpisodeInfo{},
		},
		{
			in:       "Movie.m2ts",
			wantInfo: EpisodeInfo{},
		},
		{
			in: "[爱恋&漫猫字幕组] 间谍过家家_SPY × FAMILY (01-12Fin WEBRIP 1080p AVC AAC 2022年4月 简中)",
			wantInfo: EpisodeInfo{
				Start:    episodePtr("01", 1, 0),
				End:      episodePtr("12", 12, 0),
				Complete: true,
			},
		},
		{
			in: "[DBD-Raws][命运石之门][01-24TV全集+SP+剧场版+特典映像][1080P][BDRip][HEVC-10bit][简繁日双语外挂][FLAC][MKV]",
			wantInfo: EpisodeInfo{
				Start:    episodePtr("01", 1, 0),
				End:      episodePtr("24", 24, 0),
				Complete: true,
			},
		},
		{
			in: "[酷漫404][進擊的巨人.最終季][60-75][1080P][BDrip][繁中日特效字幕(內嵌 外掛)][HEVC-10bit AAC][MKV]",
			wantInfo: EpisodeInfo{
				Start: episodePtr("60", 60, 0),
				End:   episodePtr("75", 75, 0),
			},
		},
		{
			in: "黑袍纠察队 第二季 1-2-3-4 集 The Boys Season 2 2020.English.HD1080P.x264.DD5.1.中英双字幕.ENG.CHS.taobaobt",
			wantInfo: EpisodeInfo{
				Season: new(2),
				Start:  episodePtr("1", 1, 0),
				End:    episodePtr("4", 4, 0),
			},
		},
		{
			in: "全职猎人2011[全62集][中文字幕].Hunter.x.Hunter.S01.2011.1080p.KKTV.WEB-DL.H264.AAC-ColorTV",
			wantInfo: EpisodeInfo{
				Season:   new(1),
				Start:    episodePtr("1", 1, 0),
				End:      episodePtr("62", 62, 0),
				Complete: true,
			},
		},
		{
			in: "某动画（全十二话）",
			wantInfo: EpisodeInfo{
				Start:    episodePtr("1", 1, 0),
				End:      episodePtr("12", 12, 0),
				Complete: true,
			},
		},
		{
			in: "长篇动画[全一百集]",
			wantInfo: EpisodeInfo{
				Start:    episodePtr("1", 1, 0),
				End:      episodePtr("100", 100, 0),
				Complete: true,
			},
		},
		{
			in: "[Lilith-Raws] 進擊的巨人 第四季 01-16 [Baha][WEB-DL][1080p][AVC AAC][CHT][MKV]",
			wantInfo: EpisodeInfo{
				Season: new(4),
				Start:  episodePtr("01", 1, 0),
				End:    episodePtr("16", 16, 0),
			},
		},
		{
			in: "[动漫国字幕组]★10月新番[SPY×FAMILY间谍家家酒 _ 间谍过家家][01-50(全集)][1080P][简体][MP4]",
			wantInfo: EpisodeInfo{
				Start:    episodePtr("01", 1, 0),
				End:      episodePtr("50", 50, 0),
				Complete: true,
			},
		},
		{
			in: "Show.S01E02.mkv",
			wantInfo: EpisodeInfo{
				Season:  new(1),
				Episode: episodePtr("02", 2, 0),
			},
		},
		{
			in: "Show.S01E10.5.mkv",
			wantInfo: EpisodeInfo{
				Season:  new(1),
				Episode: episodePtr("10.5", 10, 5),
			},
		},
		{
			in: "Show（EP 08）.mkv",
			wantInfo: EpisodeInfo{
				Episode: episodePtr("08", 8, 0),
			},
		},
	}

	for _, tt := range tests {
		got := ExtractEpisodeInfo(tt.in)
		if !reflect.DeepEqual(got, tt.wantInfo) {
			t.Fatalf("input=%q, got=%+v, want=%+v", tt.in, got, tt.wantInfo)
		}
	}
}

func TestEpisodeNumberOffset(t *testing.T) {
	got := EpisodeNumber{Raw: "60.5", Major: 60, Minor: 5}.Offset(-59)
	want := "1.5"
	if got != want {
		t.Fatalf("Offset got %q, want %q", got, want)
	}
}

func episodePtr(raw string, major, minor int) *EpisodeNumber {
	return &EpisodeNumber{Raw: raw, Major: major, Minor: minor}
}

func TestChineseNumberToInt(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{in: "十", want: "10", ok: true},
		{in: "二十", want: "20", ok: true},
		{in: "二十一", want: "21", ok: true},
		{in: "", want: "", ok: false},
		{in: "百", want: "", ok: false},
		{in: "十一A", want: "", ok: false},
	}

	for _, tt := range tests {
		got, ok := chineseNumberToInt(tt.in)
		if got != tt.want || ok != tt.ok {
			t.Fatalf("input=%q, got=(%q,%v), want=(%q,%v)", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}
