// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"ohurlshortener/core"
	"ohurlshortener/storage"
	"ohurlshortener/utils"
)

// ReloadUrls
//
// 从数据库中获取所有「有效」状态的短链接
// 并将其可以 key-> value 形式存入 Redis 中
func ReloadUrls() (bool, error) {

	// 多目标地址一次性装载，避免逐条短链接查询
	destsMap, err := loadDestsMap()
	if err != nil {
		log.Println(err)
	}

	//Get total count to calculate page size
	count, err := storage.GetUrlCount(0)
	if err != nil {
		log.Println(err)
		return false, utils.RaiseError("内部错误，请联系管理员")
	}

	if count > 0 {
		// query for all urls by page
		totalPageCount := (count / 100) + 1 //100 at a time

		var wg sync.WaitGroup
		for i := 1; i <= totalPageCount; i++ {
			urls, err := storage.FindAllShortUrlsByPage(i, 100)
			if err != nil {
				log.Println(err)
				continue
			}
			wg.Add(1)
			go func(urls []core.ShortUrl) {
				defer wg.Done()
				for _, url := range urls {
					if url.Valid {
						mu := memShortUrl(url.DestUrl, url.OpenType, destsMap[url.ShortUrl])
						res, err := json.Marshal(mu)
						if err != nil {
							log.Println(err)
							continue
						}
						err = storage.RedisSet4Ever(url.ShortUrl, res)
						if err != nil {
							log.Println(err)
							continue
						}
					}
				} // end of for
			}(urls)
		}
		wg.Wait()
	}
	return true, nil
}

// loadDestsMap 装载全部多目标地址并按 short_url 分组；老库未建表时返回空 map
func loadDestsMap() (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	dests, err := storage.FindAllShortUrlDests()
	if err != nil {
		return result, err
	}
	for _, d := range dests {
		if result[d.ShortUrl] == nil {
			result[d.ShortUrl] = make(map[string]string)
		}
		result[d.ShortUrl][d.Label] = d.DestUrl
	}
	return result, nil
}

// memShortUrl 组装写入 Redis 的短链接信息
func memShortUrl(destUrl string, openType core.OpenType, dests map[string]string) core.MemShortUrl {
	return core.MemShortUrl{DestUrl: destUrl, OpenType: openType, Dests: dests}
}

// Search4ShortUrl
//
// 从 Redis 中查询目标短链接是否存在
func Search4ShortUrl(shortUrl string) (url core.MemShortUrl, err error) {
	mu := core.MemShortUrl{}
	found, err := storage.RedisGetString(shortUrl)
	if err != nil {
		log.Println(err)
		return mu, utils.RaiseError("内部错误，请联系管理员")
	}
	if utils.EmptyString(found) {
		return mu, nil
	}
	return mu, json.Unmarshal([]byte(found), &mu)
}

// GetPagesShortUrls
//
// 获取分页的短链接信息；admin 返回全部，普通用户仅返回自己创建的短链接
func GetPagesShortUrls(url string, page int, size int, operator core.User) ([]core.ShortUrl, error) {
	if page < 1 || size < 1 {
		return nil, nil
	}
	allUrls, err := storage.FindPagedShortUrls(url, page, size, ownerFilter(operator))
	if err != nil {
		log.Println(err)
		return allUrls, utils.RaiseError("内部错误，请联系管理员")
	}
	if err := attachShortUrlDests(allUrls); err != nil {
		log.Println(err)
	}
	return allUrls, nil
}

// attachShortUrlDests 为短链接列表批量附加多目标地址（管理端展示用）；查询失败仅记录日志
func attachShortUrlDests(urls []core.ShortUrl) error {
	if len(urls) == 0 {
		return nil
	}
	codes := make([]string, len(urls))
	for i, u := range urls {
		codes[i] = u.ShortUrl
	}
	dests, err := storage.FindShortUrlDestsByCodes(codes)
	if err != nil {
		return err
	}
	grouped := make(map[string]map[string]string)
	for _, d := range dests {
		if grouped[d.ShortUrl] == nil {
			grouped[d.ShortUrl] = make(map[string]string)
		}
		grouped[d.ShortUrl][d.Label] = d.DestUrl
	}
	for i := range urls {
		urls[i].Dests = grouped[urls[i].ShortUrl]
	}
	return nil
}

// ErrNoPermission 无权操作该短链接（非创建者且非 admin）
var ErrNoPermission = errors.New("无权操作该短链接")

// ownerFilter 返回归属过滤条件：admin 返回 0（不过滤），普通用户返回自身 ID
func ownerFilter(operator core.User) int {
	if operator.IsAdmin {
		return 0
	}
	return operator.ID
}

// checkOperator 校验操作者权限：admin 放行，否则仅创建者可操作
func checkOperator(found core.ShortUrl, operator core.User) error {
	if operator.IsAdmin {
		return nil
	}
	if found.CreatedBy != operator.ID {
		return ErrNoPermission
	}
	return nil
}

// GenerateShortUrl
//
// 生成短链接（单目标地址，兼容旧调用方）
func GenerateShortUrl(destUrl string, memo string, openType int, operator core.User) (string, error) {
	return GenerateShortUrlWithDests(destUrl, memo, openType, nil, operator)
}

// GenerateShortUrlWithDests
//
// 生成短链接；dests 为多目标地址列表（可为空）。
// 短码基于主目标地址 destUrl 生成，其余目标地址通过 label 选择访问。
// 同一 destUrl 已生成过短码时幂等返回已有短码；不同 destUrl 哈希碰撞仍报错。
// 同一 destUrl 被其他用户创建过时返回错误（防止越权复用他人短链）。
func GenerateShortUrlWithDests(destUrl string, memo string, openType int, dests []core.ShortUrlDest, operator core.User) (string, error) {
	shortUrl, err := core.GenerateShortLink(destUrl)
	if err != nil {
		log.Println(err)
		return "", utils.RaiseError("内部错误，请联系管理员")
	}

	foundUrl, err := storage.FindShortUrl(shortUrl)
	if err != nil {
		log.Println(err)
		return "", utils.RaiseError("内部错误，请联系管理员")
	}

	if !foundUrl.IsEmpty() {
		if foundUrl.DestUrl == destUrl {
			if err := checkOperator(foundUrl, operator); err != nil {
				return "", utils.RaiseError("该目标地址已由其他用户创建")
			}
			// 同一长链接已生成过短码，幂等返回
			return shortUrl, nil
		}
		// 不同长链接哈希碰撞
		return shortUrl, utils.RaiseError(fmt.Sprintf("短链接 %s 已存在", shortUrl))
	}

	var nsMemo sql.NullString
	if !utils.EmptyString(memo) {
		nsMemo = sql.NullString{Valid: true, String: memo}
	}

	url := core.ShortUrl{
		DestUrl:   destUrl,
		ShortUrl:  shortUrl,
		CreatedAt: time.Now(),
		Valid:     true,
		Memo:      nsMemo,
		OpenType:  core.OpenType(openType),
		CreatedBy: operator.ID,
	}

	if err := storage.InsertShortUrl(url); err != nil {
		if storage.IsUniqueViolation(err) {
			existing, ferr := storage.FindShortUrl(shortUrl)
			if ferr != nil || existing.IsEmpty() {
				log.Println(ferr)
				return "", utils.RaiseError("内部错误，请联系管理员")
			}
			if cerr := checkOperator(existing, operator); cerr != nil {
				return "", utils.RaiseError("该目标地址已由其他用户创建")
			}
			// 并发创建同一长链接，唯一索引兜底，幂等返回
			return shortUrl, nil
		}
		log.Println(err)
		return "", utils.RaiseError("内部错误，请联系管理员")
	}

	if len(dests) > 0 {
		if err := storage.InsertShortUrlDests(shortUrl, dests); err != nil {
			log.Println(err)
			return "", utils.RaiseError("多目标地址写入失败，请确认已执行 sql/add_short_url_dests.sql 迁移脚本")
		}
	}

	mu := memShortUrl(url.DestUrl, url.OpenType, destsToMap(dests))
	res, err := json.Marshal(mu)
	if err != nil {
		return "", utils.RaiseError("内部错误，请联系管理员")
	}

	if err := storage.RedisSet4Ever(shortUrl, res); err != nil {
		log.Println(err)
		return "", utils.RaiseError("内部错误，请联系管理员")
	}

	return shortUrl, nil
}

// destsToMap 将多目标地址列表转换为 label -> dest_url 映射
func destsToMap(dests []core.ShortUrlDest) map[string]string {
	if len(dests) == 0 {
		return nil
	}
	result := make(map[string]string, len(dests))
	for _, d := range dests {
		result[d.Label] = d.DestUrl
	}
	return result
}

// AppendShortUrlDests
//
// 向已有短链接追加多目标地址。必须携带已有短码 shortUrl；
// label 已存在时返回错误（禁止覆盖既有目标，防止篡改），追加成功后重写该短码的 Redis 缓存。
func AppendShortUrlDests(shortUrl string, dests []core.ShortUrlDest, operator core.User) (string, error) {
	if len(dests) == 0 {
		return "", utils.RaiseError("destinations 不能为空")
	}

	found, err := storage.FindShortUrl(shortUrl)
	if err != nil {
		log.Println(err)
		return "", utils.RaiseError("内部错误，请联系管理员")
	}
	if found.IsEmpty() {
		return "", utils.RaiseError("该短链接不存在")
	}
	if err := checkOperator(found, operator); err != nil {
		return "", err
	}

	existing, err := loadShortUrlDests(shortUrl)
	if err != nil {
		log.Println(err)
		return "", utils.RaiseError("内部错误，请联系管理员")
	}
	for _, d := range dests {
		if _, ok := existing[d.Label]; ok {
			return "", utils.RaiseError(fmt.Sprintf("label %s 已存在，不允许覆盖", d.Label))
		}
	}

	if err := storage.InsertShortUrlDests(shortUrl, dests); err != nil {
		if storage.IsUniqueViolation(err) {
			// 并发追加同 label 时由唯一索引兜底
			return "", utils.RaiseError("label 已存在，不允许覆盖")
		}
		log.Println(err)
		return "", utils.RaiseError("多目标地址写入失败，请确认已执行 sql/add_short_url_dests.sql 迁移脚本")
	}

	allDests, err := loadShortUrlDests(shortUrl)
	if err != nil {
		log.Println(err)
	}

	mu := memShortUrl(found.DestUrl, found.OpenType, allDests)
	res, err := json.Marshal(mu)
	if err != nil {
		return "", utils.RaiseError("内部错误，请联系管理员")
	}
	if err := storage.RedisSet4Ever(shortUrl, res); err != nil {
		log.Println(err)
	}

	return shortUrl, nil
}

// ChangeState
//
// 禁用/启用短链接
func ChangeState(shortUrl string, enable bool, operator core.User) (bool, error) {
	found, err := storage.FindShortUrl(shortUrl)
	if err != nil {
		return false, utils.RaiseError("内部错误，请联系管理员")
	}

	if found.IsEmpty() {
		return false, utils.RaiseError("该短链接不存在")
	}

	if err := checkOperator(found, operator); err != nil {
		return false, err
	}

	found.Valid = enable

	e := storage.UpdateShortUrl(found)
	if e != nil {
		return false, utils.RaiseError("内部错误，请联系管理员")
	}

	if enable {
		dests, err := loadShortUrlDests(shortUrl)
		if err != nil {
			log.Println(err)
		}
		mu := memShortUrl(found.DestUrl, found.OpenType, dests)
		res, err := json.Marshal(mu)
		if err != nil {
			return false, utils.RaiseError("内部错误，请联系管理员")
		}
		storage.RedisSet4Ever(found.ShortUrl, res)
	} else {
		storage.RedisDelete(found.ShortUrl)
	}

	return true, nil
}

// loadShortUrlDests 查询指定短链接的多目标地址映射；无记录或查询失败时返回 nil
func loadShortUrlDests(shortUrl string) (map[string]string, error) {
	dests, err := storage.FindShortUrlDests(shortUrl)
	if err != nil {
		return nil, err
	}
	return destsToMap(dests), nil
}

// DeleteUrlAndAccessLogs 删除短链接以及对应的访问日志
func DeleteUrlAndAccessLogs(shortUrl string, operator core.User) error {
	found, err := storage.FindShortUrl(shortUrl)
	if err != nil {
		return utils.RaiseError("内部错误，请联系管理员")
	}

	if found.IsEmpty() {
		return utils.RaiseError("该短链接不存在")
	}

	if err := checkOperator(found, operator); err != nil {
		return err
	}

	err = storage.DeleteShortUrlWithAccessLogs(found)
	if err != nil {
		return utils.RaiseError("内部错误，无法删除短链接")
	}

	storage.RedisDelete(found.ShortUrl)

	return nil
}
