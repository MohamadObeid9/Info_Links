ALTER TABLE public.page_views DROP CONSTRAINT IF EXISTS page_views_device_type_chk;
ALTER TABLE public.page_views DROP COLUMN IF EXISTS device_type;
