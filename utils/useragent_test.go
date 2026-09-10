// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package utils

import (
	"testing"
)

func TestIsWeChatUA(t *testing.T) {

	ua1 := "mozilla/5.0 (linux; u; android 4.1.2; zh-cn; mi-one plus build/jzo54k) applewebkit/534.30 (khtml, like gecko) version/4.0 mobile safari/534.30 micromessenger/5.0.1.352"
	ua2 := "mozilla/5.0 (iphone; cpu iphone os 5_1_1 like mac os x) applewebkit/534.46 (khtml, like gecko) mobile/9b206 micromessenger/5.0"
	ua3 := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"

	type args struct {
		ua string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test1", args: args{ua: ua1}, want: true},
		{name: "Test2", args: args{ua: ua2}, want: true},
		{name: "Test3", args: args{ua: ua3}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWeChatUA(tt.args.ua); got != tt.want {
				t.Errorf("IsWeChatUA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSafari(t *testing.T) {
	ua1 := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
	ua2 := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Version/11 Safari/537.36"
	type args struct {
		ua string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test1", args: args{ua: ua1}, want: false},
		{name: "Test2", args: args{ua: ua2}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSafari(tt.args.ua); got != tt.want {
				t.Errorf("IsSafari() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeviceTierFromHints(t *testing.T) {
	tests := []struct {
		name                    string
		mobile, platform, model string
		want                    DeviceTier
	}{
		{name: "ipados", mobile: "?0", platform: `"iPadOS"`, want: DeviceTierTablet},
		{name: "ios iphone", mobile: "?1", platform: `"iOS"`, want: DeviceTierMobile},
		{name: "ios desktop mode", mobile: "?0", platform: `"iOS"`, want: DeviceTierTablet},
		{name: "ios model ipad", mobile: "?1", platform: `"iOS"`, model: `"iPad13,4"`, want: DeviceTierTablet},
		{name: "android", mobile: "?1", platform: `"Android"`, want: DeviceTierMobile},
		{name: "macos", mobile: "?0", platform: `"macOS"`, want: DeviceTierPC},
		{name: "macos ipad model", mobile: "?0", platform: `"macOS"`, model: `"iPad12,1"`, want: DeviceTierTablet},
		{name: "windows", mobile: "?0", platform: `"Windows"`, want: DeviceTierPC},
		{name: "chrome os", mobile: "?0", platform: `"Chrome OS"`, want: DeviceTierPC},
		{name: "mobile hint only", mobile: "?1", want: DeviceTierMobile},
		{name: "desktop hint only", mobile: "?0", want: DeviceTierPC},
		{name: "unquoted values tolerated", mobile: "1", platform: "android", want: DeviceTierMobile},
		{name: "no hints", want: DeviceTierUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeviceTierFromHints(tt.mobile, tt.platform, tt.model); got != tt.want {
				t.Errorf("DeviceTierFromHints(%q, %q, %q) = %v, want %v", tt.mobile, tt.platform, tt.model, got, tt.want)
			}
		})
	}
}

func TestDeviceTierFromUA(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want DeviceTier
	}{
		{name: "android old format", ua: "Mozilla/5.0 (Linux; Android/14; Pixel 8) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36", want: DeviceTierMobile},
		{name: "android modern format", ua: "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36", want: DeviceTierMobile},
		{name: "iphone old format", ua: "Mozilla/5.0 (iPhone/17.0; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1", want: DeviceTierMobile},
		{name: "iphone modern format", ua: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1", want: DeviceTierMobile},
		{name: "ipad old format", ua: "Mozilla/5.0 (iPad/17.0; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1", want: DeviceTierTablet},
		{name: "ipad modern format", ua: "Mozilla/5.0 (iPad; CPU OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4 Mobile/15E148 Safari/604.1", want: DeviceTierTablet},
		{name: "ipados desktop mode", ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15) AppleWebKit/605.1.15 Version/13.0 Safari/605.1.15 Mobile/15E148", want: DeviceTierTablet},
		{name: "pc", ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/120.0 Safari/537.36", want: DeviceTierPC},
		{name: "empty", ua: "", want: DeviceTierPC},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeviceTierFromUA(tt.ua); got != tt.want {
				t.Errorf("DeviceTierFromUA(%q) = %v, want %v", tt.ua, got, tt.want)
			}
		})
	}
}
