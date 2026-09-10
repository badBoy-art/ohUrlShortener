// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package core

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"ohurlshortener/utils"
)

type OpenType int

const (
	OpenInAll OpenType = iota
	OpenInWeChat
	OpenInDingTalk
	OpenInIPhone
	OpenInAndroid
	OpenInIPad
	OpenInSafari
	OpenInChrome
	OpenInFirefox
)

// 访问短链接时用于选择目标地址的请求头，均由客户端（App 等）自行设置，
// 浏览器不会携带，因此可作为区分客户端类型的可靠依据
const (
	// ClientTypeHeader 客户端类型标识，由 App 等客户端自行设置（如 app/wechat）
	ClientTypeHeader = "X-Client-Type"
	// PlatformHeader 客户端平台标识，由 App 等客户端自行设置（如 android/ios/ipad）
	PlatformHeader = "X-Platform"
)

// 无上述请求头（如浏览器访问）时，根据 User-Agent 自动识别设备类型所用的标识
const (
	LabelPC     = "pc"
	LabelMobile = "mobile"
	LabelTablet = "tablet"
)

// 浏览器 Client Hints 请求头（Chromium 系浏览器在站点声明 Accept-CH 后发送，
// Safari 暂不支持），优先用于设备类型识别，User-Agent 识别作为兜底
const (
	// SecCHUAMobileHeader 是否为移动设备（?1 / ?0）
	SecCHUAMobileHeader = "Sec-CH-UA-Mobile"
	// SecCHUAPlatformHeader 平台标识（如 "iOS"、"Android"、"macOS"）
	SecCHUAPlatformHeader = "Sec-CH-UA-Platform"
	// SecCHUAModelHeader 设备型号（如 "iPhone 15"、"Pixel 8"）
	SecCHUAModelHeader = "Sec-CH-UA-Model"
)

// Accept-CH 响应头：声明本服务需要的 Client Hints，
// 浏览器收到后会在后续请求中携带 Sec-CH-UA-* 请求头
const (
	AcceptCHHeader = "Accept-CH"
	AcceptCHValue  = "Sec-CH-UA-Mobile, Sec-CH-UA-Platform, Sec-CH-UA-Model"
)

// 多目标地址数量与字段长度限制
const (
	MaxDestCount    = 20
	MaxDestLabelLen = 64
	MaxDestUrlLen   = 2048
)

// ShortUrl 短链接
type ShortUrl struct {
	ID        int64          `db:"id"`
	ShortUrl  string         `db:"short_url"`
	DestUrl   string         `db:"dest_url"`
	CreatedAt time.Time      `db:"created_at"`
	Valid     bool           `db:"is_valid"`
	Memo      sql.NullString `db:"memo"`
	OpenType  OpenType       `db:"open_type"`
	// Dests 多目标地址（label -> dest_url），仅用于管理端展示，不落库
	Dests map[string]string `db:"-"`
}

// ShortUrlDest 多目标短链接中的一条目标地址，label 为访问时用于选择的标识
type ShortUrlDest struct {
	ID        int64     `db:"id" json:"id"`
	ShortUrl  string    `db:"short_url" json:"short_url"`
	Label     string    `db:"label" json:"label"`
	DestUrl   string    `db:"dest_url" json:"dest_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type MemShortUrl struct {
	DestUrl  string            `json:"dest_url"`
	OpenType OpenType          `json:"open_type"`
	Dests    map[string]string `json:"dests,omitempty"`
}

// ResolveDestUrl 按优先级依次尝试给定的目标标识，命中即返回对应目标地址；
// 全部未命中（或标识为空）时回退到主目标地址
func (m MemShortUrl) ResolveDestUrl(labels ...string) string {
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if utils.EmptyString(label) {
			continue
		}
		if dest, ok := m.Dests[label]; ok && !utils.EmptyString(dest) {
			return dest
		}
	}
	return m.DestUrl
}

// ValidateShortUrlDests 校验多目标地址列表：数量、label 与 dest_url 非空及长度、label 唯一
func ValidateShortUrlDests(dests []ShortUrlDest) error {
	if len(dests) == 0 {
		return nil
	}
	if len(dests) > MaxDestCount {
		return fmt.Errorf("destinations 最多 %d 个", MaxDestCount)
	}
	labels := make(map[string]bool, len(dests))
	for _, d := range dests {
		if utils.EmptyString(d.Label) {
			return fmt.Errorf("destinations 中 label 不能为空")
		}
		if len(strings.TrimSpace(d.Label)) > MaxDestLabelLen {
			return fmt.Errorf("destinations 中 label 长度不能超过 %d", MaxDestLabelLen)
		}
		if utils.EmptyString(d.DestUrl) {
			return fmt.Errorf("destinations 中 label %s 的 dest_url 不能为空", d.Label)
		}
		if len(strings.TrimSpace(d.DestUrl)) > MaxDestUrlLen {
			return fmt.Errorf("destinations 中 dest_url 长度不能超过 %d", MaxDestUrlLen)
		}
		if labels[d.Label] {
			return fmt.Errorf("destinations 中 label %s 重复", d.Label)
		}
		labels[d.Label] = true
	}
	return nil
}

// IsEmpty 判断是否为空
func (url ShortUrl) IsEmpty() bool {
	return reflect.DeepEqual(url, ShortUrl{})
}

// GenerateShortLink 生成短链接
func GenerateShortLink(initialLink string) (string, error) {
	if utils.EmptyString(initialLink) {
		return "", fmt.Errorf("empty string")
	}
	urlHash, err := utils.Sha256Of(initialLink)
	if err != nil {
		return "", err
	}
	// number := new(big.Int).SetBytes(urlHash).Uint64()
	// str := utils.Base58Encode([]byte(fmt.Sprintf("%d", number)))
	str := utils.Base58Encode(urlHash)
	return str[:8], nil
}
