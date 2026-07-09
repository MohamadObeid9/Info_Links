ALTER TABLE public.link_clicks
  DROP CONSTRAINT IF EXISTS link_clicks_one_target_chk;

ALTER TABLE public.link_clicks
  DROP CONSTRAINT IF EXISTS link_clicks_extra_link_id_fkey;

ALTER TABLE public.link_clicks
  DROP COLUMN IF EXISTS extra_link_id;
