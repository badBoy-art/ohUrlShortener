package core

import (
	"strings"
	"testing"
)

func TestGenerateShortLink(t *testing.T) {
	link1, err := GenerateShortLink("https://www.example.com/very/long/path")
	if err != nil {
		t.Fatal(err)
	}
	link2, err := GenerateShortLink("https://www.example.com/very/long/path")
	if err != nil {
		t.Fatal(err)
	}
	if len(link1) != 8 {
		t.Errorf("expected length 8, got %d", len(link1))
	}
	if link1 != link2 {
		t.Errorf("same input should produce same short link: %s != %s", link1, link2)
	}
	if _, err := GenerateShortLink(""); err == nil {
		t.Error("empty input should fail")
	}
}

func TestMemShortUrlResolveDestUrl(t *testing.T) {
	mu := MemShortUrl{
		DestUrl: "https://primary.example.com",
		Dests: map[string]string{
			"pc":     "https://pc.example.com",
			"mobile": "https://m.example.com",
		},
	}
	tests := []struct {
		name  string
		label string
		want  string
	}{
		{name: "hit pc", label: "pc", want: "https://pc.example.com"},
		{name: "hit mobile", label: "mobile", want: "https://m.example.com"},
		{name: "miss label falls back to primary", label: "ipad", want: "https://primary.example.com"},
		{name: "empty label falls back to primary", label: "", want: "https://primary.example.com"},
		{name: "label with spaces is trimmed", label: "  pc  ", want: "https://pc.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mu.ResolveDestUrl(tt.label); got != tt.want {
				t.Errorf("ResolveDestUrl(%q) = %q, want %q", tt.label, got, tt.want)
			}
		})
	}

	// 老数据：无多目标地址，任何 label 都回退主目标
	legacy := MemShortUrl{DestUrl: "https://legacy.example.com"}
	if got := legacy.ResolveDestUrl("pc"); got != "https://legacy.example.com" {
		t.Errorf("legacy ResolveDestUrl() = %q, want primary", got)
	}
}

func TestMemShortUrlResolveDestUrlPriority(t *testing.T) {
	mu := MemShortUrl{
		DestUrl: "https://primary.example.com",
		Dests: map[string]string{
			"app":    "https://app.example.com",
			"ios":    "https://ios.example.com",
			"pc":     "https://pc.example.com",
			"mobile": "https://m.example.com",
		},
	}
	tests := []struct {
		name   string
		labels []string
		want   string
	}{
		{name: "first label wins", labels: []string{"app", "ios", "mobile"}, want: "https://app.example.com"},
		{name: "skip miss then hit", labels: []string{"ipad", "ios", "mobile"}, want: "https://ios.example.com"},
		{name: "skip empty labels", labels: []string{"", "  ", "mobile"}, want: "https://m.example.com"},
		{name: "explicit over ua", labels: []string{"pc", "mobile"}, want: "https://pc.example.com"},
		{name: "all miss falls back to primary", labels: []string{"wechat", "android"}, want: "https://primary.example.com"},
		{name: "no labels falls back to primary", labels: nil, want: "https://primary.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mu.ResolveDestUrl(tt.labels...); got != tt.want {
				t.Errorf("ResolveDestUrl(%v) = %q, want %q", tt.labels, got, tt.want)
			}
		})
	}
}

func TestValidateShortUrlDests(t *testing.T) {
	valid := []ShortUrlDest{
		{Label: "pc", DestUrl: "https://pc.example.com"},
		{Label: "mobile", DestUrl: "https://m.example.com"},
	}
	if err := ValidateShortUrlDests(valid); err != nil {
		t.Errorf("valid dests should pass: %v", err)
	}
	if err := ValidateShortUrlDests(nil); err != nil {
		t.Errorf("nil dests should pass: %v", err)
	}

	tests := []struct {
		name  string
		dests []ShortUrlDest
		err   string
	}{
		{name: "empty label", dests: []ShortUrlDest{{Label: "", DestUrl: "https://a.example.com"}}, err: "label 不能为空"},
		{name: "empty dest_url", dests: []ShortUrlDest{{Label: "pc", DestUrl: ""}}, err: "dest_url 不能为空"},
		{name: "duplicate label", dests: []ShortUrlDest{{Label: "pc", DestUrl: "https://a.example.com"}, {Label: "pc", DestUrl: "https://b.example.com"}}, err: "重复"},
		{name: "too many", dests: make([]ShortUrlDest, MaxDestCount+1), err: "最多"},
		{name: "label too long", dests: []ShortUrlDest{{Label: strings.Repeat("l", MaxDestLabelLen+1), DestUrl: "https://a.example.com"}}, err: "label 长度"},
		{name: "dest_url too long", dests: []ShortUrlDest{{Label: "pc", DestUrl: "https://a.example.com/" + strings.Repeat("x", MaxDestUrlLen)}}, err: "dest_url 长度"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShortUrlDests(tt.dests)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.err)
			}
			if !strings.Contains(err.Error(), tt.err) {
				t.Errorf("expected error containing %q, got %q", tt.err, err.Error())
			}
		})
	}
}
