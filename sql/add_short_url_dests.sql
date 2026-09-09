-- Migration: multi-destination short links
--
-- Run this script against an EXISTING oh_url_shortener database to upgrade
-- it in place. Old data (short_urls.dest_url) keeps working: it remains the
-- primary destination of every link, and rows in this table are only added
-- when a link is created with multiple destinations.
--
--   psql -U postgres -d oh_url_shortener -f sql/add_short_url_dests.sql

CREATE TABLE IF NOT EXISTS public.short_url_dests (
  id serial4 NOT NULL,
  short_url varchar(200) NOT NULL,
  label varchar(64) NOT NULL,
  dest_url text NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT short_url_dests_pk PRIMARY KEY (id),
  CONSTRAINT short_url_dests_un UNIQUE (short_url, label)
);
CREATE INDEX IF NOT EXISTS short_url_dests_short_url_idx ON public.short_url_dests (short_url);
