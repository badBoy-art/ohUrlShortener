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
	"fmt"
	"log"
	"time"

	"ohurlshortener/core"
	"ohurlshortener/storage"
	"ohurlshortener/utils"
)

const accessLogsPrefix = "OH_ACCESS_LOGS#"

// NewAccessLog 记录访问日志
func NewAccessLog(url string, ip string, useragent string, referer string) error {
	var (
		l = core.AccessLog{
			ShortUrl:   url,
			AccessTime: time.Now(),
			Ip:         sql.NullString{String: ip, Valid: true},
			UserAgent:  sql.NullString{String: useragent, Valid: true},
		}
		logJson, _ = json.Marshal(l)
		key        = fmt.Sprintf("%s%s", accessLogsPrefix, utils.UserAgentIpHash(useragent, ip))
		err        = storage.RedisSet30m(key, logJson)
	)

	if err != nil {
		log.Println(err)
		return utils.RaiseError("内部错误，请联系管理员")
	}

	return nil
}

// StoreAccessLogs 将访问日志存入数据库
func StoreAccessLogs() error {
	keys, err := storage.RedisScan4Keys(accessLogsPrefix + "*")
	if err != nil {
		log.Println(err)
		return utils.RaiseError("内部错误，请联系管理员")
	}

	var logs []core.AccessLog
	for _, k := range keys {
		v, err := storage.RedisGetString(k)
		if err != nil {
			log.Printf("redis error for key %s", k)
			continue
		}
		accessLog := core.AccessLog{}
		json.Unmarshal([]byte(v), &accessLog)
		logs = append(logs, accessLog)
	} // end of for

	err = storage.InsertAccessLogs(logs)
	if err != nil {
		log.Println(err)
		return utils.RaiseError("内部错误，请联系管理员")
	}

	err = storage.RedisDelete(keys...)
	if err != nil {
		log.Println(err)
		return utils.RaiseError("内部错误，请联系管理员")
	}
	return nil
}

// GetPagedAccessLogs 获取分页访问日志；admin 返回全部，普通用户仅自己名下短链的日志
func GetPagedAccessLogs(url string, start, end string, page, size int, operator core.User) ([]core.AccessLog, error) {
	if page < 1 || size < 1 {
		return nil, nil
	}
	allAccessLogs, err := storage.FindAllAccessLogs(url, start, end, page, size, ownerFilter(operator))
	if err != nil {
		log.Println(err)
		return allAccessLogs, utils.RaiseError("内部错误，请联系管理员")
	}
	return allAccessLogs, nil
}

// GetAccessLogsCount 获取访问日志总数；admin 统计全部，普通用户仅自己名下短链的日志
func GetAccessLogsCount(url string, start, end string, operator core.User) (int, int, error) {
	return storage.FindAccessLogsCount(url, start, end, ownerFilter(operator))
}

// GetAllAccessLogs 获取所有访问日志；admin 返回全部，普通用户仅自己名下短链的日志
func GetAllAccessLogs(url string, operator core.User) ([]core.AccessLog, error) {
	allAccessLogs, err := storage.FindAllAccessLogsByUrl(url, ownerFilter(operator))
	if err != nil {
		log.Println(err)
		return allAccessLogs, utils.RaiseError("内部错误，请联系管理员")
	}
	return allAccessLogs, nil
}
