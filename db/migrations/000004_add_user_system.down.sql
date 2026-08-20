DROP INDEX IF EXISTS public.contributions_user_id_created_at_idx;
ALTER TABLE public.contributions DROP CONSTRAINT IF EXISTS contributions_user_id_fkey;
ALTER TABLE public.contributions DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS public.reports_user_id_created_at_idx;
ALTER TABLE public.reports DROP CONSTRAINT IF EXISTS reports_user_id_fkey;
ALTER TABLE public.reports DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS public.feedback_user_id_created_at_idx;
ALTER TABLE public.feedback DROP CONSTRAINT IF EXISTS feedback_user_id_fkey;
ALTER TABLE public.feedback DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS public.link_clicks_user_id_clicked_at_idx;
ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_user_id_fkey;
ALTER TABLE public.link_clicks DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS public.page_views_user_id_visited_at_idx;
ALTER TABLE public.page_views DROP CONSTRAINT IF EXISTS page_views_user_id_fkey;
ALTER TABLE public.page_views DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS public.favorite_events_user_id_created_at_idx;
DROP TABLE IF EXISTS public.favorite_events;

DROP INDEX IF EXISTS public.users_unique_username;
DROP TABLE IF EXISTS public.users;
