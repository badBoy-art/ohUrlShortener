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
	// 兼容新旧格式：Android/4.1.2（旧）与 Android 14（新）
	regex := regexp.MustCompile(`(?i)Android[\/ ][\d.]+`)
	return regex.MatchString(ua)
}

// IsHarmonyOS 判断是否为华为鸿蒙系统（HarmonyOS 2-4 与 HarmonyOS NEXT 均携带相关标识）
func IsHarmonyOS(ua string) bool {
	regex := regexp.MustCompile(`(?i)(HarmonyOS|OpenHarmony)`)
	return regex.MatchString(ua)
}

// harmonyDeviceKind 解析鸿蒙 UA 首段的设备形态标识：
// (Phone; HarmonyOS 5.0) / (Tablet; OpenHarmony 5.0) / (PC; HarmonyOS 5.0)
func harmonyDeviceKind(ua string) (DeviceTier, bool) {
	regex := regexp.MustCompile(`(?i)\((Phone|Tablet|PC)\s*;\s*(?:HarmonyOS|OpenHarmony)[\/ ]?[\d.]+\)`)
	m := regex.FindStringSubmatch(ua)
	if len(m) < 2 {
		return DeviceTierUnknown, false
	}
	switch strings.ToLower(m[1]) {
	case "tablet":
		return DeviceTierTablet, true
	case "pc":
		return DeviceTierPC, true
	default:
		return DeviceTierMobile, true
	}
}

// androidTablet 判断 Android 系（含 MIUI / 澎湃 HyperOS）平板：
// 带显式 Tablet/Pad 系列标识，或 Chrome 系 UA 无 Mobile 标记（Android 平板 Chrome 不带 Mobile）
func androidTablet(ua string) bool {
	if !IsAndroid(ua) {
		return false
	}
	if regexp.MustCompile(`(?i)\b(Tablet|MediaPad|MIPAD|MI PAD|MatePad)\b`).MatchString(ua) {
		return true
	}
	lower := strings.ToLower(ua)
	return strings.Contains(lower, "chrome") && !strings.Contains(lower, "mobile")
}

func IsIPhone(ua string) bool {
	// 兼容新旧格式：iPhone/12.1（旧）与 iPhone; CPU iPhone OS 17_5（新）
	regex := regexp.MustCompile(`(?i)iPhone[\/;]`)
	return regex.MatchString(ua)
}

func IsIPad(ua string) bool {
	// 兼容新旧格式：iPad/17.0（旧）与 iPad; CPU OS 17_5（新）
	regex := regexp.MustCompile(`(?i)iPad[\/;]`)
	return regex.MatchString(ua)
}

func IsTablet(ua string) bool {
	// iPad 原生 UA（新旧格式），或 iPadOS 13+ 桌面模式 UA（Macintosh 平台 + Mobile 标记）
	regex := regexp.MustCompile(`(?i)(iPad[\/;]|Macintosh.*Mobile\/[\d.]+)`)
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
		// Client Hints 无法区分 Android 平板与手机（均上报 ?1），
		// UA 回退时按是否带 Mobile 标记识别平板
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
	if tier, ok := harmonyDeviceKind(ua); ok {
		return tier
	}
	if androidTablet(ua) {
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
