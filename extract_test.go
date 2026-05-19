package episodex

import "testing"

func TestExtract(t *testing.T) {
	tests := []struct {
		in     string
		season string
		ep     string
	}{
		{in: "My.Show.S01E02.1080p.WEB-DL.mkv", season: "01", ep: "02"},
		{in: "某动画 第2季 第10集.mp4", season: "02", ep: "10"},
		{in: "番剧标题 [12 END][1080p].mkv", season: "", ep: "12"},
		{in: "Anime - EP07 - x265.mkv", season: "", ep: "07"},
		{in: "Anime.10v1.1080p.mkv", season: "", ep: "10"},
		{in: "Anime.10.5v1.BDRip.mkv", season: "", ep: "10.5"},
		{in: "S02", season: "02", ep: ""},
		{in: "S02E03.mkv", season: "02", ep: "03"},
		{in: "2020", season: "", ep: ""},
		{in: "第一季", season: "01", ep: ""},
		{in: "第十一季 100", season: "11", ep: "100"},
		{in: "示例剧.第一百季.1080p.mkv", season: "", ep: ""},
	}

	for _, tt := range tests {
		season, ep := Extract(tt.in)
		if season != tt.season || ep != tt.ep {
			t.Fatalf("input=%q, got season=%q ep=%q, want season=%q ep=%q",
				tt.in, season, ep, tt.season, tt.ep)
		}
	}
}

func TestExtractSingleField(t *testing.T) {
	if got := ExtractSeason("Show S03 1080p.mkv"); got != "03" {
		t.Fatalf("ExtractSeason got %q, want %q", got, "03")
	}
	if got := ExtractEpisode("Show [EP12].mkv"); got != "12" {
		t.Fatalf("ExtractEpisode got %q, want %q", got, "12")
	}
	if got := ExtractEpisode("Show.S02E03.mkv"); got != "03" {
		t.Fatalf("ExtractEpisode got %q, want %q", got, "03")
	}
	if got := ExtractEpisode("S02.mkv"); got != "" {
		t.Fatalf("ExtractEpisode got %q, want empty", got)
	}
	if got := ExtractSeason(""); got != "" {
		t.Fatalf("ExtractSeason got %q, want empty", got)
	}
	if got := ExtractEpisode(".mkv"); got != "" {
		t.Fatalf("ExtractEpisode got %q, want empty", got)
	}
	if got := ExtractEpisode("OVA.3.5.mkv"); got != "3.5" {
		t.Fatalf("ExtractEpisode got %q, want %q", got, "3.5")
	}
	if got := ExtractEpisode("Show.2019.BluRay.mkv"); got != "" {
		t.Fatalf("ExtractEpisode got %q, want empty", got)
	}
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
