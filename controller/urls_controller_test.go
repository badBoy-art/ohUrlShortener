package controller

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"ohurlshortener/core"

	"github.com/gin-gonic/gin"
)

func TestDeviceLabel(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{name: "android", ua: "Mozilla/5.0 (Linux; Android/14; Pixel 8) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36", want: core.LabelMobile},
		{name: "iphone", ua: "Mozilla/5.0 (iPhone/17.0; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1", want: core.LabelMobile},
		{name: "ipad", ua: "Mozilla/5.0 (iPad/17.0; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1", want: core.LabelTablet},
		{name: "ipados desktop mode", ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0 Safari/605.1.15 Mobile/15E148", want: core.LabelTablet},
		{name: "wechat on android", ua: "Mozilla/5.0 (Linux; Android/13; SM-G9910) MicroMessenger/8.0.40 NetType/WIFI", want: core.LabelMobile},
		{name: "pc chrome", ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", want: core.LabelPC},
		{name: "pc edge", ua: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0", want: core.LabelPC},
		{name: "empty ua", ua: "", want: core.LabelPC},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deviceLabel(tt.ua); got != tt.want {
				t.Errorf("deviceLabel(%q) = %q, want %q", tt.ua, got, tt.want)
			}
		})
	}
}

func TestDestLabels(t *testing.T) {
	newCtx := func(headers map[string]string, ua string) *gin.Context {
		req := httptest.NewRequest("GET", "/abc123", nil)
		req.Header.Set("User-Agent", ua)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = req
		return ctx
	}

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
