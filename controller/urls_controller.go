// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package controller

import (
	"net/http"

	"ohurlshortener/core"
	"ohurlshortener/service"
	"ohurlshortener/utils"

	"github.com/gin-gonic/gin"
)

// ShortUrlDetail 重定向到目标地址
func ShortUrlDetail(c *gin.Context) {
	url := c.Param("url")
	if utils.EmptyString(url) {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "404 - ohUrlShortener",
			"code":    http.StatusNotFound,
			"message": "您访问的页面已失效",
			"label":   "Status Not Found",
		})
		return
	}

	memUrl, err := service.Search4ShortUrl(url)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"title":   "内部错误 - ohUrlShortener",
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
			"label":   "Error",
		})
		return
	}

	if utils.EmptyString(memUrl.DestUrl) {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title":   "404 - ohUrlShortener",
			"code":    http.StatusNotFound,
			"message": "您访问的页面已失效",
			"label":   "Status Not Found",
		})
		return
	}

	// 目标地址选择优先级：
	// App 客户端类型 X-Client-Type > App 平台 X-Platform > User-Agent 自动识别（pc/mobile/tablet） > 主目标地址
	destUrl := memUrl.ResolveDestUrl(destLabels(c)...)

	ua := c.Request.UserAgent()
	switch ot := memUrl.OpenType; ot {
	case core.OpenInAndroid:
		if utils.IsAndroid(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInDingTalk:
		if utils.IsDingTalk(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInChrome:
		if utils.IsChrome(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInIPad:
		if utils.IsIPad(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInIPhone:
		if utils.IsIPhone(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInSafari:
		if utils.IsSafari(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInWeChat:
		if utils.IsWeChatUA(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInFirefox:
		if utils.IsFirefox(ua) {
			redirectSuccess(url, destUrl, c)
		} else {
			redirectFail(c)
		}
	case core.OpenInAll:
		redirectSuccess(url, destUrl, c)
	}
}

func redirectSuccess(shortUrl, destUrl string, ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, destUrl)
	go service.NewAccessLog(shortUrl, ctx.ClientIP(), ctx.Request.UserAgent(), ctx.Request.Referer())
}

// destLabels 组装目标地址标识的匹配优先级：
// X-Client-Type > X-Platform > User-Agent 自动识别（pc/mobile/tablet）
// 前两者由 App 等客户端自行设置（浏览器不会携带），命中即返回对应目标地址
func destLabels(c *gin.Context) []string {
	labels := []string{}
	if v := c.GetHeader(core.ClientTypeHeader); !utils.EmptyString(v) {
		labels = append(labels, v)
	}
	if v := c.GetHeader(core.PlatformHeader); !utils.EmptyString(v) {
		labels = append(labels, v)
	}
	labels = append(labels, deviceLabel(c.Request.UserAgent()))
	return labels
}

// deviceLabel 根据浏览器默认携带的 User-Agent 识别设备类型，返回三档标识：
// tablet（iPad 及 iPadOS 13+ 桌面模式）/ mobile（Android、iPhone）/ pc（其余）
func deviceLabel(ua string) string {
	if utils.IsTablet(ua) {
		return core.LabelTablet
	}
	if utils.IsAndroid(ua) || utils.IsIPhone(ua) {
		return core.LabelMobile
	}
	return core.LabelPC
}

func redirectFail(ctx *gin.Context) {
	ctx.HTML(http.StatusNotFound, "error.html", gin.H{
		"title":   "404 - ohUrlShortener",
		"code":    http.StatusNotFound,
		"message": "不支持的打开方式",
		"label":   "Status Not Found",
	})
}
