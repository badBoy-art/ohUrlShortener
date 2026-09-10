// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package utils

import (
	"regexp"
	"strings"
)

// DeviceTier 设备类型三档：pc / mobile / tablet
type DeviceTier int

const (
	DeviceTierUnknown DeviceTier = iota
	DeviceTierPC
	DeviceTierMobile
	DeviceTierTablet
)

func IsAndroid(ua string) bool {
	regex := regexp.MustCompile(`(?i)Android\/[\d.]+`)
	return regex.MatchString(ua)
}

func IsIPhone(ua string) bool {
	regex := regexp.MustCompile(`(?i)iPhone\/[\d.]+`)
	return regex.MatchString(ua)
}

func IsIPad(ua string) bool {
	regex := regexp.MustCompile(`(?i)iPad\/[\d.]+`)
	return regex.MatchString(ua)
}

func IsTablet(ua string) bool {
	// iPad 原生 UA，或 iPadOS 13+ 桌面模式 UA（Macintosh 平台 + Mobile 标记）
	regex := regexp.MustCompile(`(?i)(iPad\/[\d.]+|Macintosh.*Mobile\/[\d.]+)`)
	return regex.MatchString(ua)
}

func IsWeChatUA(ua string) bool {
	regex := regexp.MustCompile(`(?i)MicroMessenger\/[\d.]+`)
	return regex.MatchString(ua)
}

func IsDingTalk(ua string) bool {
	regex := regexp.MustCompile(`(?i)DingTalk\/[\d.]+`)
	return regex.MatchString(ua)
}

func IsSafari(ua string) bool {
	regex := regexp.MustCompile(`(?i)Version\/[\d.]+ Safari\/[\d.]+`)
	return regex.MatchString(ua)
}

func IsChrome(ua string) bool {
	regex := regexp.MustCompile(`(?i)Chrome\/[\d.]+ Safari`)
	return regex.MatchString(ua)
}

func IsFirefox(ua string) bool {
	regex := regexp.MustCompile(`(?i)Firefox\/[\d.]+`)
	return regex.MatchString(ua)
}

// DeviceTierFromHints 根据浏览器 Client Hints 请求头识别设备类型三档；
// 信息不足无法判断时返回 DeviceTierUnknown，调用方应回退到 User-Agent 识别
func DeviceTierFromHints(mobileHint, platformHint, modelHint string) DeviceTier {
	platform := unquote(strings.ToLower(strings.TrimSpace(platformHint)))
	model := strings.ToLower(unquote(strings.TrimSpace(modelHint)))
	mobile := strings.TrimSpace(mobileHint)

	isMobile := mobile == "?1" || mobile == "1" || strings.EqualFold(mobile, "true")
	isDesktop := mobile == "?0" || mobile == "0" || strings.EqualFold(mobile, "false")

	switch platform {
	case "ipados":
		return DeviceTierTablet
	case "ios":
		// iPad 请求桌面模式时 Sec-CH-UA-Mobile 为 ?0
		if isDesktop || strings.Contains(model, "ipad") {
			return DeviceTierTablet
		}
		return DeviceTierMobile
	case "android":
		// Android 平板与手机的 UA / Client Hints 均无法区分，平板应用请走 X-Platform 请求头
		return DeviceTierMobile
	case "chrome os", "chromium os", "macos", "windows", "linux":
		// iPadOS 13+ 桌面模式可能上报 macOS，结合型号再判断一次
		if strings.Contains(model, "ipad") {
			return DeviceTierTablet
		}
		return DeviceTierPC
	}
	if isMobile {
		return DeviceTierMobile
	}
	if isDesktop {
		return DeviceTierPC
	}
	return DeviceTierUnknown
}

// DeviceTierFromUA 根据 User-Agent 正则识别设备类型三档，识别不出时归为 pc
func DeviceTierFromUA(ua string) DeviceTier {
	if IsTablet(ua) {
		return DeviceTierTablet
	}
	if IsAndroid(ua) || IsIPhone(ua) {
		return DeviceTierMobile
	}
	return DeviceTierPC
}

func unquote(s string) string {
	return strings.Trim(s, `"`)
}
