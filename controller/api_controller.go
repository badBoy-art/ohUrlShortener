// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ohurlshortener/core"
	"ohurlshortener/service"
	"ohurlshortener/utils"

	"github.com/gin-gonic/gin"
)

// APINewAdmin
//
// Add new admin user
func APINewAdmin(ctx *gin.Context) {
	account := ctx.PostForm("account")
	password := ctx.PostForm("password")
	if utils.EmptyString(account) || utils.EmptyString(password) {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("用户名或密码不能为空"))
		return
	}

	if len(password) < 8 {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("密码长度最少8位"))
		return
	}

	err := service.NewUser(account, password)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, core.ResultJsonSuccess())
}

// APIAdminUpdate
//
// Update password of given admin user
func APIAdminUpdate(ctx *gin.Context) {
	account := ctx.Param("account")
	password := ctx.PostForm("password")

	if utils.EmptyString(account) || utils.EmptyString(password) {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("用户名或密码不能为空"))
		return
	}

	if len(password) < 8 {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("密码长度最少8位"))
		return
	}

	err := service.UpdatePassword(strings.TrimSpace(account), password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("修改失败"))
		return
	}

	ctx.JSON(http.StatusOK, core.ResultJsonSuccess())
}

// APIGenShortUrl Generate new short url
func APIGenShortUrl(ctx *gin.Context) {
	url := strings.TrimSpace(ctx.PostForm("dest_url"))
	memo := strings.TrimSpace(ctx.PostForm("memo"))
	strOpenType := ctx.PostForm("open_type")
	openType, err := strconv.Atoi(strOpenType)
	if err != nil {
		openType = int(core.OpenInAll)
	}

	dests, err := parseDestinations(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest(err.Error()))
		return
	}

	// 携带已有短码：向该短链接追加多目标地址（dest_url 可省略）
	shortUrl := strings.TrimSpace(ctx.PostForm("short_url"))
	if !utils.EmptyString(shortUrl) {
		res, err := service.AppendShortUrlDests(shortUrl, dests)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest(err.Error()))
			return
		}
		ctx.JSON(http.StatusOK, core.ResultJsonSuccessWithData(map[string]string{
			"short_url": fmt.Sprintf("%s%s", utils.AppConfig.UrlPrefix, res),
		}))
		return
	}

	// dest_url 未提供时，取第一个目标地址作为主目标（用于生成短码与默认跳转）
	if utils.EmptyString(url) && len(dests) > 0 {
		url = dests[0].DestUrl
	}

	if utils.EmptyString(url) {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("dest_url 不能为空（或提供 destinations 参数）"))
		return
	}

	res, err := service.GenerateShortUrlWithDests(url, memo, openType, dests)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest(err.Error()))
		return
	}

	json := map[string]string{
		"short_url": fmt.Sprintf("%s%s", utils.AppConfig.UrlPrefix, res),
	}
	ctx.JSON(http.StatusOK, core.ResultJsonSuccessWithData(json))
}

// parseDestinations 解析多目标地址参数 destinations（JSON 数组形式）
//
//	示例：destinations=[{"label":"pc","dest_url":"https://www.example.com"},{"label":"mobile","dest_url":"https://m.example.com"}]
func parseDestinations(ctx *gin.Context) ([]core.ShortUrlDest, error) {
	raw := strings.TrimSpace(ctx.PostForm("destinations"))
	if utils.EmptyString(raw) {
		return nil, nil
	}
	var items []struct {
		Label   string `json:"label"`
		DestUrl string `json:"dest_url"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, utils.RaiseError("destinations 参数不是合法的 JSON 数组")
	}
	dests := make([]core.ShortUrlDest, 0, len(items))
	for _, it := range items {
		dests = append(dests, core.ShortUrlDest{
			Label:   strings.TrimSpace(it.Label),
			DestUrl: strings.TrimSpace(it.DestUrl),
		})
	}
	if err := core.ValidateShortUrlDests(dests); err != nil {
		return nil, err
	}
	return dests, nil
}

// APIUrlInfo Get Short Url Stat Info.
func APIUrlInfo(ctx *gin.Context) {
	url := ctx.Param("url")
	if utils.EmptyString(strings.TrimSpace(url)) {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("url 不能为空"))
		return
	}

	stat, err := service.GetShortUrlStats(strings.TrimSpace(url))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, core.ResultJsonError(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, core.ResultJsonSuccessWithData(stat))
}

// APIUpdateUrl Enable or Disable Short Url
func APIUpdateUrl(ctx *gin.Context) {
	url := ctx.Param("url")
	enableStr := ctx.PostForm("enable")
	if utils.EmptyString(strings.TrimSpace(url)) {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("url 不能为空"))
		return
	}

	enable, err := strconv.ParseBool(enableStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("enable 参数值非法"))
		return
	}

	res, err := service.ChangeState(url, enable)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, core.ResultJsonSuccessWithData(res))
}

// APIDeleteUrl Delete Short Url
func APIDeleteUrl(ctx *gin.Context) {
	url := ctx.Param("url")
	if utils.EmptyString(strings.TrimSpace(url)) {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest("url 不能为空"))
		return
	}
	err := service.DeleteUrlAndAccessLogs(url)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, core.ResultJsonBadRequest(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, core.ResultJsonSuccess)
}
