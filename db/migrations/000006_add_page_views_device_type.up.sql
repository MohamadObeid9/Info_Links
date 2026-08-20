-- Coarse phone vs laptop classification on each visit, derived server-side from
-- User-Agent. Nullable so existing rows stay valid and are omitted from splits.

ALTER TABLE public.page_views ADD COLUMN device_type text;

ALTER TABLE public.page_views
  ADD CONSTRAINT page_views_device_type_chk
    CHECK (device_type IS NULL OR device_type = ANY (ARRAY['phone'::text, 'laptop'::text]));
