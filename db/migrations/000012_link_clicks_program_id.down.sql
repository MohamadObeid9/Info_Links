DROP INDEX IF EXISTS public.link_clicks_program_id_idx;
ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_program_id_fkey;
ALTER TABLE public.link_clicks DROP COLUMN IF EXISTS program_id;
