-- programs
ALTER TABLE public.programs ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.programs_id_seq;
ALTER TABLE public.programs ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.programs', 'id'), COALESCE((SELECT max(id) FROM public.programs), 1), true);

-- years
ALTER TABLE public.years ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.years_id_seq;
ALTER TABLE public.years ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.years', 'id'), COALESCE((SELECT max(id) FROM public.years), 1), true);

-- semesters
ALTER TABLE public.semesters ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.semesters_id_seq;
ALTER TABLE public.semesters ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.semesters', 'id'), COALESCE((SELECT max(id) FROM public.semesters), 1), true);

-- courses
ALTER TABLE public.courses ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.courses_id_seq;
ALTER TABLE public.courses ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.courses', 'id'), COALESCE((SELECT max(id) FROM public.courses), 1), true);

-- links
ALTER TABLE public.links ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.links_id_seq;
ALTER TABLE public.links ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.links', 'id'), COALESCE((SELECT max(id) FROM public.links), 1), true);

-- extra_sections
ALTER TABLE public.extra_sections ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.extra_sections_id_seq;
ALTER TABLE public.extra_sections ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.extra_sections', 'id'), COALESCE((SELECT max(id) FROM public.extra_sections), 1), true);

-- extra_links
ALTER TABLE public.extra_links ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.extra_links_id_seq;
ALTER TABLE public.extra_links ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.extra_links', 'id'), COALESCE((SELECT max(id) FROM public.extra_links), 1), true);

-- reports
ALTER TABLE public.reports ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.reports_id_seq;
ALTER TABLE public.reports ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.reports', 'id'), COALESCE((SELECT max(id) FROM public.reports), 1), true);

-- contributions
ALTER TABLE public.contributions ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.contributions_id_seq;
ALTER TABLE public.contributions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.contributions', 'id'), COALESCE((SELECT max(id) FROM public.contributions), 1), true);

ALTER TABLE public.page_views ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.page_views_id_seq;
ALTER TABLE public.page_views ALTER COLUMN id TYPE integer USING id::integer;
ALTER TABLE public.page_views ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.page_views', 'id'), COALESCE((SELECT max(id) FROM public.page_views), 1), true);


ALTER TABLE public.feedback ALTER COLUMN id DROP IDENTITY IF EXISTS;
ALTER TABLE public.feedback ALTER COLUMN id TYPE integer USING id::integer;
ALTER TABLE public.feedback ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.feedback', 'id'), COALESCE((SELECT max(id) FROM public.feedback), 1), true);

DO $$
DECLARE
  v_conname text;
BEGIN
  SELECT c.conname
    INTO v_conname
  FROM pg_constraint c
  WHERE c.conrelid = 'public.feedback'::regclass
    AND c.contype = 'c'
    AND pg_get_constraintdef(c.oid) ILIKE '%status%'
  LIMIT 1;

  IF v_conname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE public.feedback DROP CONSTRAINT %I', v_conname);
  END IF;
END $$;

ALTER TABLE public.feedback
  ALTER COLUMN category TYPE text,
  ALTER COLUMN status TYPE text;

ALTER TABLE public.feedback
ALTER COLUMN status SET DEFAULT 'new';

ALTER TABLE public.feedback
  ADD CONSTRAINT feedback_status_check CHECK (status IN ('new', 'read'));



ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_link_id_fkey;
ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_extra_link_id_fkey;

ALTER TABLE public.link_clicks ALTER COLUMN id DROP IDENTITY IF EXISTS;
ALTER TABLE public.link_clicks ALTER COLUMN id TYPE integer USING id::integer;
ALTER TABLE public.link_clicks ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.link_clicks', 'id'), COALESCE((SELECT max(id) FROM public.link_clicks), 1), true);

ALTER TABLE public.link_clicks ALTER COLUMN link_id TYPE integer USING link_id::integer;
ALTER TABLE public.link_clicks ALTER COLUMN extra_link_id TYPE integer USING extra_link_id::integer;

ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_link_id_fkey
    FOREIGN KEY (link_id) REFERENCES public.links(id);
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_extra_link_id_fkey
    FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id);
