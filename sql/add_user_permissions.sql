-- Migration: user-level permissions (created_by ownership + admin role)
--
-- Run this script against an EXISTING oh_url_shortener database to upgrade
-- it in place. Existing short_urls rows are assigned to the admin account
-- (ohUrlShortener); only admins can operate them afterwards.
--
--   psql -U postgres -d oh_url_shortener -f sql/add_user_permissions.sql

ALTER TABLE public.users ADD COLUMN IF NOT EXISTS is_admin bool NOT NULL DEFAULT false;

ALTER TABLE public.short_urls ADD COLUMN IF NOT EXISTS created_by int4;

-- Assign existing rows to the admin account (fallback: lowest user id)
UPDATE public.short_urls SET created_by = COALESCE(
  (SELECT u.id FROM public.users u WHERE u.account = 'ohUrlShortener'),
  (SELECT min(u.id) FROM public.users u)
) WHERE created_by IS NULL;

ALTER TABLE public.short_urls ALTER COLUMN created_by SET NOT NULL;

ALTER TABLE public.short_urls DROP CONSTRAINT IF EXISTS short_urls_created_by_fk;
ALTER TABLE public.short_urls ADD CONSTRAINT short_urls_created_by_fk
  FOREIGN KEY (created_by) REFERENCES public.users(id);
CREATE INDEX IF NOT EXISTS short_urls_created_by_idx ON public.short_urls (created_by);

UPDATE public.users SET is_admin = true WHERE account = 'ohUrlShortener';
