// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package service

import (
	"ohurlshortener/core"
	"ohurlshortener/storage"
	"ohurlshortener/utils"
)

// GetSumOfUrlStats 获取所有短链接的统计信息；admin 返回全部，普通用户仅统计自己名下
func GetSumOfUrlStats(operator core.User) (int, core.ShortUrlStats, error) {
	var (
		totalCount int
		result     core.ShortUrlStats
	)

	totalCount, err := storage.GetUrlCount(ownerFilter(operator))
	if err != nil {
		return totalCount, result, utils.RaiseError("内部错误，请联系管理员！")
	}

	result, er := storage.GetSumOfUrlStats(ownerFilter(operator))
	if er != nil {
		return totalCount, result, utils.RaiseError("内部错误，请联系管理员！")
	}

	return totalCount, result, nil
}

// GetShortUrlStats 获取单个短链接的统计信息（仅创建者或 admin 可查）
func GetShortUrlStats(url string, operator core.User) (core.ShortUrlStats, error) {
	found, err := storage.FindShortUrl(url)
	if err != nil {
		return core.ShortUrlStats{}, utils.RaiseError("内部错误，请联系管理员！")
	}
	if found.IsEmpty() {
		return core.ShortUrlStats{}, utils.RaiseError("该短链接不存在")
	}
	if err := checkOperator(found, operator); err != nil {
		return core.ShortUrlStats{}, err
	}
	stats, err := storage.GetUrlStats(url)
	if err != nil {
		return stats, utils.RaiseError("内部错误，请联系管理员！")
	}
	return stats, nil
}

// GetTop25Url 获取访问量最高的 25 个短链接；admin 返回全部，普通用户仅自己名下
func GetTop25Url(operator core.User) ([]core.Top25Url, error) {
	found, err := storage.GetTop25(ownerFilter(operator))
	if err != nil {
		return found, utils.RaiseError("内部错误，请联系管理员！")
	}
	return found, nil
}

// GetPagedUrlIpCountStats 获取单个短链接的 IP 访问量统计信息；admin 返回全部，普通用户仅自己名下
func GetPagedUrlIpCountStats(url string, page int, size int, operator core.User) ([]core.UrlIpCountStats, error) {
	if page < 1 || size < 1 {
		return nil, nil
	}
	found, err := storage.FindPagedUrlIpCountStats(url, page, size, ownerFilter(operator))
	if err != nil {
		return found, utils.RaiseError("内部错误，请联系管理员！")
	}
	return found, nil
}
