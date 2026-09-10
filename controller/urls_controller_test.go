package controller

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"ohurlshortener/core"

	"github.com/gin-gonic/gin"
)

func newCtx(headers map[string]string, ua string) *gin.Context {
	req := httptest.NewRequest("GET", "/abc123", nil)
	req.Header.Set("User-Agent", ua)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req
	return ctx
}

func TestDeviceLabel(t *testing.T) {
	pcUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
	mobileUA := "Mozilla/5.0 (Linux; Android/14) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36"
	tabletUA := "Mozilla/5.0 (iPad/17.0; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1"

	tests := []struct {
		name    string
		headers map[string]string
		ua      string
		want    string
	}{
		// Client Hints 优先
		{name: "ch ipados", headers: map[string]string{"Sec-CH-UA-Mobile": "?0", "Sec-CH-UA-Platform": `"iPadOS"`}, ua: pcUA, want: core.LabelTablet},
		{name: "ch ios iphone", headers: map[string]string{"Sec-CH-UA-Mobile": "?1", "Sec-CH-UA-Platform": `"iOS"`}, ua: mobileUA, want: core.LabelMobile},
		{name: "ch ios desktop mode", headers: map[string]string{"Sec-CH-UA-Mobile": "?0", "Sec-CH-UA-Platform": `"iOS"`}, ua: tabletUA, want: core.LabelTablet},
		{name: "ch ios model ipad", headers: map[string]string{"Sec-CH-UA-Mobile": "?1", "Sec-CH-UA-Platform": `"iOS"`, "Sec-CH-UA-Model": `"iPad13,4"`}, ua: mobileUA, want: core.LabelTablet},
		{name: "ch android", headers: map[string]string{"Sec-CH-UA-Mobile": "?1", "Sec-CH-UA-Platform": `"Android"`}, ua: mobileUA, want: core.LabelMobile},
		{name: "ch macos", headers: map[string]string{"Sec-CH-UA-Mobile": "?0", "Sec-CH-UA-Platform": `"macOS"`}, ua: pcUA, want: core.LabelPC},
		{name: "ch macos ipad model", headers: map[string]string{"Sec-CH-UA-Mobile": "?0", "Sec-CH-UA-Platform": `"macOS"`, "Sec-CH-UA-Model": `"iPad12,1"`}, ua: pcUA, want: core.LabelTablet},
		{name: "ch windows", headers: map[string]string{"Sec-CH-UA-Mobile": "?0", "Sec-CH-UA-Platform": `"Windows"`}, ua: pcUA, want: core.LabelPC},
		{name: "ch mobile hint only", headers: map[string]string{"Sec-CH-UA-Mobile": "?1"}, ua: pcUA, want: core.LabelMobile},
		{name: "ch desktop hint only", headers: map[string]string{"Sec-CH-UA-Mobile": "?0"}, ua: pcUA, want: core.LabelPC},
		// User-Agent 兜底
		{name: "ua android", ua: mobileUA, want: core.LabelMobile},
		{name: "ua iphone", ua: "Mozilla/5.0 (iPhone/17.0; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1", want: core.LabelMobile},
		{name: "ua ipad", ua: tabletUA, want: core.LabelTablet},
		{name: "ua ipados desktop mode", ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15) AppleWebKit/605.1.15 Version/13.0 Safari/605.1.15 Mobile/15E148", want: core.LabelTablet},
		{name: "ua pc", ua: pcUA, want: core.LabelPC},
		{name: "empty everything", want: core.LabelPC},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deviceLabel(newCtx(tt.headers, tt.ua)); got != tt.want {
				t.Errorf("deviceLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDestLabels(t *testing.T) {
	pcUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0 Safari/537.36"
	mobileUA := "Mozilla/5.0 (Linux; Android/14) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36"
	tabletUA := "Mozilla/5.0 (iPad/17.0; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1"

	tests := []struct {
		name    string
		headers map[string]string
		ua      string
		want    []string
	}{
		{name: "browser pc only", ua: pcUA, want: []string{core.LabelPC}},
		{name: "browser mobile only", ua: mobileUA, want: []string{core.LabelMobile}},
		{name: "browser tablet only", ua: tabletUA, want: []string{core.LabelTablet}},
		{name: "browser client hints only", headers: map[string]string{"Sec-CH-UA-Mobile": "?0", "Sec-CH-UA-Platform": `"iPadOS"`}, ua: pcUA, want: []string{core.LabelTablet}},
		{name: "app client type and platform first", headers: map[string]string{"X-Client-Type": "app", "X-Platform": "ios"}, ua: mobileUA, want: []string{"app", "ios", core.LabelMobile}},
		{name: "client type only", headers: map[string]string{"X-Client-Type": "wechat"}, ua: mobileUA, want: []string{"wechat", core.LabelMobile}},
		{name: "platform only", headers: map[string]string{"X-Platform": "ipad"}, ua: tabletUA, want: []string{"ipad", core.LabelTablet}},
		{name: "empty headers ignored", headers: map[string]string{"X-Client-Type": "  ", "X-Platform": ""}, ua: pcUA, want: []string{core.LabelPC}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := destLabels(newCtx(tt.headers, tt.ua))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("destLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
