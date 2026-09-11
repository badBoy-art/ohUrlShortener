-- Database Structure For ohUrlShortener
CREATE DATABASE oh_url_shortener ENCODING 'UTF8';

-- Connect to database oh_url_shortener
\c oh_url_shortener

CREATE TABLE public.short_urls (
  id serial4 NOT NULL,
	short_url varchar(200) NOT NULL,
	dest_url text NOT NULL,	
	created_at timestamp with time zone NOT NULL DEFAULT now(),
	is_valid bool NOT NULL DEFAULT true,	
	memo text,
	open_type int8 NOT NULL DEFAULT 0,
	created_by int4 NOT NULL,
	CONSTRAINT short_urls_pk PRIMARY KEY (id),
	CONSTRAINT short_urls_un UNIQUE (short_url)
);

-- Multi-destination table: one short_url can map to multiple dest_urls,
-- each identified by a label. dest_url in short_urls remains the primary
-- destination for links created before this feature (old data).
CREATE TABLE public.short_url_dests (
  id serial4 NOT NULL,
  short_url varchar(200) NOT NULL,
  label varchar(64) NOT NULL,
  dest_url text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT short_url_dests_pk PRIMARY KEY (id),
  CONSTRAINT short_url_dests_un UNIQUE (short_url, label)
);
CREATE INDEX short_url_dests_short_url_idx ON public.short_url_dests (short_url);

CREATE TABLE public.access_logs (
	id serial4 NOT NULL,
	short_url varchar(200) NOT NULL,
	access_time timestamp with time zone NOT NULL DEFAULT NOW(),
	ip varchar(64) NULL,
	user_agent varchar(1000) NULL,
	CONSTRAINT access_logs_pk PRIMARY KEY (id)
);
CREATE INDEX access_logs_short_url_idx ON public.access_logs (short_url);
CREATE INDEX access_logs_access_time_idx ON public.access_logs (access_time);
CREATE INDEX access_logs_ip_idx ON public.access_logs (ip);
CREATE INDEX access_logs_ua_idx ON public.access_logs (user_agent);

CREATE TABLE public.users (
  id serial4 NOT NULL,
	account varchar(200) NOT NULL,
	password text NOT NULL,			
	created_at timestamp with time zone NOT NULL DEFAULT NOW(), 
	is_enable bool NOT NULL DEFAULT true,	 
	is_admin bool NOT NULL DEFAULT false,
	CONSTRAINT users_pk PRIMARY KEY (id),
	CONSTRAINT users_account_un UNIQUE (account)
);

-- account: ohUrlShortener password: -2aDzm=0(ln_9^1
INSERT INTO public.users (account, "password", is_admin) VALUES('ohUrlShortener', 'EZ2zQjC3fqbkvtggy9p2YaJiLwx1kKPTJxvqVzowtx6t', true);

ALTER TABLE public.short_urls ADD CONSTRAINT short_urls_created_by_fk
  FOREIGN KEY (created_by) REFERENCES public.users(id);
CREATE INDEX short_urls_created_by_idx ON public.short_urls (created_by);

-- Insert new data
INSERT INTO public.short_urls(short_url, dest_url, created_at, is_valid, memo,open_type, created_by) VALUES
	('AC7VgPE9', 'https://www.gitlink.org.cn/baladiwei/ohurlshortener', NOW(), true, '短链接系统 gitlink 页面',0, (SELECT id FROM public.users WHERE account = 'ohUrlShortener')),
	('AvTkHZP7', 'https://gitee.com/barat/ohurlshortener', NOW(), true, '短链接系统 gitee 页面',0, (SELECT id FROM public.users WHERE account = 'ohUrlShortener')),
	('gkT39tb5', 'https://github.com/barats/ohUrlShortener', NOW(), true, '短链接系统 github 页面',0, (SELECT id FROM public.users WHERE account = 'ohUrlShortener')),
	('9HtCr7YN', 'https://www.ohurls.cn', NOW(), true, 'ohUrlShortener 短链接系统首页',0, (SELECT id FROM public.users WHERE account = 'ohUrlShortener'));


-- Create table for top25 urls
CREATE TABLE public.stats_top25 (
	id serial4 NOT NULL,
	short_url varchar(200) NOT NULL,
	today_count int8 NOT NULL DEFAULT 0,
	d_today_count int8 NOT NULL DEFAULT 0,
	stats_time timestamp with time zone NOT NULL DEFAULT NOW(), 
	CONSTRAINT stats_tv_pk PRIMARY KEY (id)
);

-- Stored procedure for top25 urls 
CREATE FUNCTION p_stats_top25() RETURNS void AS $$
BEGIN
	RAISE NOTICE 'Procedure p_stats_top25() called';
	
	-- delete all records 
	DELETE FROM public.stats_top25 WHERE 1=1;

	-- insert fresh-new records
	INSERT INTO public.stats_top25(short_url,today_count,d_today_count,stats_time)  
		SELECT l.short_url AS short_url, COUNT(l.ip) AS today_count ,COUNT(DISTINCT(l.ip)) AS d_today_count, NOW() AS stats_time
		FROM public.access_logs l WHERE date(l.access_time) = date(NOW()) GROUP BY l.short_url ORDER BY today_count DESC LIMIT 25;
END; 
$$ LANGUAGE plpgsql;

-- Create table for sum view 
CREATE TABLE public.stats_sum (
	stats_key varchar(200) NOT NULL,
	stats_value int8 NOT NULL DEFAULT 0,
	CONSTRAINT stats_sum_key PRIMARY KEY (stats_key)
);

-- Insert pre-defined stats 
INSERT INTO public.stats_sum (stats_key,stats_value) VALUES  
	('today_count',0), ('d_today_count',0),
	('yesterday_count',0), ('d_yesterday_count',0),
	('last_7_days_count',0), ('d_last_7_days_count',0),
	('monthly_count',0), ('d_monthly_count',0);

-- Stored procedure for stats sum view 
CREATE FUNCTION p_stats_sum() RETURNS void AS $$ 
DECLARE 
	today_count int8;
	d_today_count int8;
	yesterday_count int8;
	d_yesterday_count int8;
	last_7_days_count int8;
	d_last_7_days_count int8;
	monthly_count int8;
	d_monthly_count int8;
BEGIN
	RAISE NOTICE 'Procedure p_stats_sum() called';
	
	SELECT COUNT(l.ip),COUNT(DISTINCT(l.ip)) INTO today_count,d_today_count 
	FROM public.access_logs l WHERE date(l.access_time) = date(NOW());

	SELECT COUNT(l.ip),COUNT(DISTINCT(l.ip)) INTO yesterday_count,d_yesterday_count 
	FROM public.access_logs l WHERE date(l.access_time) = (NOW() - INTERVAL '1 day')::date;

	SELECT COUNT(l.ip),COUNT(DISTINCT(l.ip)) INTO last_7_days_count,d_last_7_days_count 
	FROM public.access_logs l WHERE date(l.access_time) >= (NOW() - INTERVAL '7 day')::date;

	SELECT COUNT(l.ip),COUNT(DISTINCT(l.ip)) INTO monthly_count,d_monthly_count 
	FROM public.access_logs l WHERE DATE_PART('month', l.access_time) = DATE_PART('month',NOW());

	UPDATE public.stats_sum SET stats_value = 
	CASE
		WHEN stats_key = 'today_count' THEN today_count
		WHEN stats_key = 'd_today_count' THEN d_today_count
		WHEN stats_key = 'yesterday_count' THEN yesterday_count
		WHEN stats_key = 'd_yesterday_count' THEN d_yesterday_count
		WHEN stats_key = 'last_7_days_count' THEN last_7_days_count
		WHEN stats_key = 'd_last_7_days_count' THEN d_last_7_days_count
		WHEN stats_key = 'monthly_count' THEN monthly_count
		WHEN stats_key = 'd_monthly_count' THEN d_monthly_count
		ELSE 0
	END;	
END;
$$ LANGUAGE plpgsql;

-- Create table for ip url sum 
CREATE TABLE public.stats_ip_sum (
	short_url varchar(200) NOT NULL,
	today_count int8 NOT NULL DEFAULT 0,
	d_today_count int8 NOT NULL DEFAULT 0,
	yesterday_count int8 NOT NULL DEFAULT 0,
	d_yesterday_count int8 NOT NULL DEFAULT 0,
	last_7_days_count int8 NOT NULL DEFAULT 0,
	d_last_7_days_count int8 NOT NULL DEFAULT 0,
	monthly_count int8 NOT NULL DEFAULT 0,
	d_monthly_count int8 NOT NULL DEFAULT 0,
	total_count int8 NOT NULL DEFAULT 0,
	d_total_count int8 NOT NULL DEFAULT 0,
	CONSTRAINT stats_ip_sum_pk PRIMARY KEY (short_url)	
);

-- Stored procedure for ip url sum 
CREATE FUNCTION p_stats_ip_sum() RETURNS void AS $$ 
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