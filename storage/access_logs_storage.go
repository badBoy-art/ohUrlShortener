// Copyright (c) [2022] [巴拉迪维 BaratSemet]
// [ohUrlShortener] is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
// 				 http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package storage

import (
	"fmt"

	"ohurlshortener/core"
	"ohurlshortener/utils"
)

func DeleteAccessLogs(shortUrl string) error {
	query := `DELETE from public.access_logs WHERE short_url = :short_url`
	return DbNamedExec(query, shortUrl)
}

func FindAccessLogs(shortUrl string) ([]core.AccessLog, error) {
	var (
		found []core.AccessLog
		query = "SELECT * FROM public.access_logs l WHERE l.short_url = $1 ORDER BY l.id DESC"
		err   = DbSelect(query, &found, shortUrl)
	)

	return found, err
}

func InsertAccessLogs(logs []core.AccessLog) error {
	if len(logs) <= 0 {
		return nil
	}
	query := `INSERT INTO public.access_logs (short_url, access_time, ip, user_agent) VALUES(:short_url,:access_time,:ip,:user_agent)`
	logsSlice := splitLogsArray(logs, MaxInsertCount)
	for _, slice := range logsSlice {
		if err := DbNamedExec(query, slice); err != nil {
			return err
		}
	}
	return nil
}

// FindAccessLogsCount
//
// # Find Access Logs Count and Unique IP Count
//
// First return value is total_count, Second return value is unique_ip_count ip count
//
// ownerID > 0 时仅统计该用户短链接的访问日志；ownerID = 0 统计全部
func FindAccessLogsCount(url string, start, end string, ownerID int) (int, int, error) {
	type LogsCount struct {
		TotalCount    int `db:"total_count"`
		UniqueIpCount int `db:"unique_ip_count"`
	}
	from := ` FROM public.access_logs l WHERE 1=1 `
	if ownerID > 0 {
		from = ` FROM public.access_logs l JOIN public.short_urls u ON l.short_url = u.short_url WHERE 1=1 `
	}
	query := `SELECT count(l.id) as total_count, count(distinct(l.ip)) as unique_ip_count` + from
	args := []interface{}{}
	if !utils.EmptyString(url) {
		query += fmt.Sprintf(` AND l.short_url = $%d`, len(args)+1)
		args = append(args, url)
	}
	if !utils.EmptyString(start) {
		query += fmt.Sprintf(` AND l.access_time >= to_date($%d,'YYYY-MM-DD')`, len(args)+1)
		args = append(args, start)
	}
	if !utils.EmptyString(end) {
		query += fmt.Sprintf(` AND l.access_time < to_date($%d,'YYYY-MM-DD')`, len(args)+1)
		args = append(args, end)
	}
	if ownerID > 0 {
		query += fmt.Sprintf(` AND u.created_by = $%d`, len(args)+1)
		args = append(args, ownerID)
	}
	var count LogsCount
	return count.TotalCount, count.UniqueIpCount, DbGet(query, &count, args...)
}

func FindAllAccessLogs(url string, start, end string, page, size int, ownerID int) ([]core.AccessLog, error) {
	var (
		found      []core.AccessLog
		offset     = (page - 1) * size
		selectCols = `SELECT *`
		from       = ` FROM public.access_logs l WHERE 1=1 `
		query      string
		args       = []interface{}{}
	)

	if ownerID > 0 {
		selectCols = `SELECT l.*`
		from = ` FROM public.access_logs l JOIN public.short_urls u ON l.short_url = u.short_url WHERE 1=1 `
	}
	query = selectCols + from

	if !utils.EmptyString(url) {
		query += fmt.Sprintf(` AND l.short_url = $%d`, len(args)+1)
		args = append(args, url)
	}
	if !utils.EmptyString(start) {
		query += fmt.Sprintf(` AND l.access_time >= to_date($%d,'YYYY-MM-DD')`, len(args)+1)
		args = append(args, start)
	}
	if !utils.EmptyString(end) {
		query += fmt.Sprintf(` AND l.access_time < to_date($%d,'YYYY-MM-DD')`, len(args)+1)
		args = append(args, end)
	}
	if ownerID > 0 {
		query += fmt.Sprintf(` AND u.created_by = $%d`, len(args)+1)
		args = append(args, ownerID)
	}

	query += fmt.Sprintf(` ORDER BY l.id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, size, offset)
	return found, DbSelect(query, &found, args...)
}

func FindAllAccessLogsByUrl(url string, ownerID int) ([]core.AccessLog, error) {
	found := []core.AccessLog{}
	selectCols := `SELECT *`
	from := ` FROM public.access_logs l WHERE 1=1 `
	if ownerID > 0 {
		selectCols = `SELECT l.*`
		from = ` FROM public.access_logs l JOIN public.short_urls u ON l.short_url = u.short_url WHERE 1=1 `
	}
	query := selectCols + from
	args := []interface{}{}
	if !utils.EmptyString(url) {
		query += fmt.Sprintf(` AND l.short_url = $%d`, len(args)+1)
		args = append(args, url)
	}
	if ownerID > 0 {
		query += fmt.Sprintf(` AND u.created_by = $%d`, len(args)+1)
		args = append(args, ownerID)
	}
	query += ` ORDER BY l.id DESC`
	return found, DbSelect(query, &found, args...)
}
