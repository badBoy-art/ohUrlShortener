package storage

import (
	"fmt"

	"ohurlshortener/core"
	"ohurlshortener/utils"
)

// GetUrlStats 获取短链接的访问量统计信息
func GetUrlStats(url string) (core.ShortUrlStats, error) {
	found := core.ShortUrlStats{}
	query := `select * from public.stats_ip_sum WHERE short_url = $1`
	err := DbGet(query, &found, url)
	return found, err
}

// GetUrlCount 获取短链接总数
//
// ownerID > 0 时仅统计该用户创建的短链接；ownerID = 0 统计全部
func GetUrlCount(ownerID int) (int, error) {
	var result int
	query := `SELECT count(l.id) FROM public.short_urls l`
	if ownerID > 0 {
		query = `SELECT count(l.id) FROM public.short_urls l WHERE l.created_by = $1`
		return result, DbGet(query, &result, ownerID)
	}
	return result, DbGet(query, &result)
}

// GetSumOfUrlStats 获取所有短链接的访问量统计信息
//
// ownerID > 0 时按归属实时聚合（stats_sum 为全库滚动表，无法过滤）；ownerID = 0 走 stats_sum
func GetSumOfUrlStats(ownerID int) (core.ShortUrlStats, error) {
	result := core.ShortUrlStats{}
	if ownerID > 0 {
		query := `
		SELECT
			COUNT(*) FILTER (WHERE date(l.access_time) = date(NOW())) AS today_count,
			COUNT(DISTINCT l.ip) FILTER (WHERE date(l.access_time) = date(NOW())) AS d_today_count,
			COUNT(*) FILTER (WHERE date(l.access_time) = (NOW() - INTERVAL '1 day')::date) AS yesterday_count,
			COUNT(DISTINCT l.ip) FILTER (WHERE date(l.access_time) = (NOW() - INTERVAL '1 day')::date) AS d_yesterday_count,
			COUNT(*) FILTER (WHERE date(l.access_time) >= (NOW() - INTERVAL '7 day')::date) AS last_7_days_count,
			COUNT(DISTINCT l.ip) FILTER (WHERE date(l.access_time) >= (NOW() - INTERVAL '7 day')::date) AS d_last_7_days_count,
			COUNT(*) FILTER (WHERE DATE_PART('month', l.access_time) = DATE_PART('month', NOW())) AS monthly_count,
			COUNT(DISTINCT l.ip) FILTER (WHERE DATE_PART('month', l.access_time) = DATE_PART('month', NOW())) AS d_monthly_count
		FROM public.access_logs l
		JOIN public.short_urls u ON l.short_url = u.short_url
		WHERE u.created_by = $1`
		return result, DbGet(query, &result, ownerID)
	}

	query := `SELECT * FROM public.stats_sum`
	data := []core.StatsSum{}
	err := DbSelect(query, &data)
	if err != nil {
		return result, err
	}
	for _, v := range data {
		switch v.Key {
		case "today_count":
			result.TodayCount = v.Value
		case "yesterday_count":
			result.YesterdayCount = v.Value
		case "last_7_days_count":
			result.Last7DaysCount = v.Value
		case "monthly_count":
			result.MonthlyCount = v.Value
		case "d_today_count":
			result.DistinctTodayCount = v.Value
		case "d_yesterday_count":
			result.DistinctYesterdayCount = v.Value
		case "d_last_7_days_count":
			result.DistinctLast7DaysCount = v.Value
		case "d_monthly_count":
			result.DistinctMonthlyCount = v.Value
		}
	}
	return result, nil
}

// GetTop25 获取访问量前 25 的短链接
//
// ownerID > 0 时仅返回该用户创建的短链接；ownerID = 0 返回全部
func GetTop25(ownerID int) ([]core.Top25Url, error) {
	query := `SELECT u.*,s.today_count AS today_count,s.d_today_count AS d_today_count FROM public.short_urls u , public.stats_top25 s WHERE u.short_url = s.short_url`
	found := []core.Top25Url{}
	if ownerID > 0 {
		query = `SELECT u.*,s.today_count AS today_count,s.d_today_count AS d_today_count FROM public.short_urls u , public.stats_top25 s WHERE u.short_url = s.short_url AND u.created_by = $1`
		return found, DbSelect(query, &found, ownerID)
	}
	return found, DbSelect(query, &found)
}

// FindPagedUrlIpCountStats 获取单个短链接的 IP 访问量统计信息
//
// ownerID > 0 时仅返回该用户创建的短链接；ownerID = 0 不过滤
func FindPagedUrlIpCountStats(url string, page int, size int, ownerID int) ([]core.UrlIpCountStats, error) {
	found := []core.UrlIpCountStats{}
	offset := (page - 1) * size
	query := `SELECT s.*,u.id,u.dest_url,u.created_at,u.is_valid,u.memo FROM public.stats_ip_sum s , public.short_urls u WHERE u.short_url = s.short_url`
	args := []interface{}{}
	if ownerID > 0 {
		query += fmt.Sprintf(` AND u.created_by = $%d`, len(args)+1)
		args = append(args, ownerID)
	}
	query += fmt.Sprintf(` ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, size, offset)
	if !utils.EmptyString(url) {
		query := `SELECT s.*,u.id,u.dest_url,u.created_at,u.is_valid,u.memo
		FROM public.stats_ip_sum s , public.short_urls u WHERE u.short_url = s.short_url AND u.short_url = $1`
		urlArgs := []interface{}{url}
		if ownerID > 0 {
			query += fmt.Sprintf(` AND u.created_by = $%d`, len(urlArgs)+1)
			urlArgs = append(urlArgs, ownerID)
		}
		query += fmt.Sprintf(` ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d`, len(urlArgs)+1, len(urlArgs)+2)
		urlArgs = append(urlArgs, size, offset)
		var foundUrl core.UrlIpCountStats
		err := DbGet(query, &foundUrl, urlArgs...)
		if !foundUrl.IsEmpty() {
			found = append(found, foundUrl)
		}
		return found, err
	}
	return found, DbSelect(query, &found, args...)
}

// CallProcedureStatsIPSum
// Call scheduled procedures to calculate stats result.
//
// Suggested time interval to call this procedure : 30 ~ 60 minutes
func CallProcedureStatsIPSum() error {
	query := `SELECT 1 AS r FROM p_stats_ip_sum()`
	var r int
	return DbGet(query, &r)
}

// CallProcedureStatsTop25
// Call scheduled procedures to calculate stats result.
//
// Suggested time interval to call this procedure 5 ~ 10 minutes
func CallProcedureStatsTop25() error {
	query := `SELECT 2 AS r FROM p_stats_top25()`
	var r int
	return DbGet(query, &r)
}

// CallProcedureStatsSum
// Call scheduled procedures to calculate stats result.
//
// Suggested time interval to call this procedure : 5 ~ 10 minutes
func CallProcedureStatsSum() error {
	query := `SELECT 3 AS r FROM p_stats_sum()`
	var r int
	return DbGet(query, &r)
}
