-- Migration: rewrite p_stats_ip_sum() as a two-level aggregation
--   The original implementation ran 10 correlated subqueries per short_url
--   (O(short_urls * access_logs)), which never finishes on large tables and
--   blocks other writes to stats_ip_sum.
--   A single-pass COUNT(...) FILTER rewrite still used COUNT(DISTINCT x) FILTER (...),
--   which PostgreSQL plans as per-group subplans and is equally slow.
--   This version aggregates first per (short_url, ip) and then per short_url,
--   so no DISTINCT or DISTINCT+FILTER aggregate is needed. Semantics are unchanged.
--   psql -U postgres -d oh_url_shortener -f sql/fix_stats_ip_sum_perf.sql

CREATE OR REPLACE FUNCTION p_stats_ip_sum() RETURNS void AS $$
BEGIN
	RAISE NOTICE 'Procedure p_stats_ip_sum() called';

	-- Delete all records
	DELETE FROM public.stats_ip_sum WHERE 1=1;

	-- Calculate new stats data
	INSERT INTO public.stats_ip_sum(short_url,today_count,d_today_count,yesterday_count,d_yesterday_count,last_7_days_count,d_last_7_days_count,
		monthly_count,d_monthly_count,total_count,d_total_count)
		SELECT
			u.short_url,
			COALESCE(g.today_count,0), COALESCE(g.d_today_count,0),
			COALESCE(g.yesterday_count,0), COALESCE(g.d_yesterday_count,0),
			COALESCE(g.last_7_days_count,0), COALESCE(g.d_last_7_days_count,0),
			COALESCE(g.monthly_count,0), COALESCE(g.d_monthly_count,0),
			COALESCE(g.total_count,0), COALESCE(g.d_total_count,0)
		FROM public.short_urls u
			LEFT JOIN (
				SELECT
					short_url,
					SUM(total)      AS total_count,
					COUNT(*) FILTER (WHERE total > 0)      AS d_total_count,
					SUM(today)      AS today_count,
					COUNT(*) FILTER (WHERE today > 0)      AS d_today_count,
					SUM(yesterday)  AS yesterday_count,
					COUNT(*) FILTER (WHERE yesterday > 0)  AS d_yesterday_count,
					SUM(last7)      AS last_7_days_count,
					COUNT(*) FILTER (WHERE last7 > 0)      AS d_last_7_days_count,
					SUM(monthly)    AS monthly_count,
					COUNT(*) FILTER (WHERE monthly > 0)    AS d_monthly_count
				FROM (
					SELECT
						short_url,
						ip,
						COUNT(*) AS total,
						COUNT(*) FILTER (WHERE date(access_time) = date(NOW())) AS today,
						COUNT(*) FILTER (WHERE date(access_time) = (NOW() - INTERVAL '1 day')::date) AS yesterday,
						COUNT(*) FILTER (WHERE date(access_time) >= (NOW() - INTERVAL '7 day')::date) AS last7,
						COUNT(*) FILTER (WHERE DATE_PART('month', access_time) = DATE_PART('month', NOW())) AS monthly
					FROM public.access_logs
					GROUP BY short_url, ip
				) p
				GROUP BY short_url
			) g ON g.short_url = u.short_url;
END;
$$ LANGUAGE plpgsql;
