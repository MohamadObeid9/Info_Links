
ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_link_id_fkey;
ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_extra_link_id_fkey;

ALTER TABLE public.link_clicks ALTER COLUMN id DROP IDENTITY IF EXISTS;
ALTER TABLE public.link_clicks ALTER COLUMN id TYPE bigint USING id::bigint;
ALTER TABLE public.link_clicks ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.link_clicks', 'id'), COALESCE((SELECT max(id) FROM public.link_clicks), 1), true);

ALTER TABLE public.link_clicks ALTER COLUMN link_id TYPE bigint USING link_id::bigint;
ALTER TABLE public.link_clicks ALTER COLUMN extra_link_id TYPE bigint USING extra_link_id::bigint;

ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_link_id_fkey
    FOREIGN KEY (link_id) REFERENCES public.links(id);
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_extra_link_id_fkey
    FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id);

ALTER TABLE public.feedback DROP CONSTRAINT IF EXISTS feedback_status_check;

ALTER TABLE public.feedback
  ALTER COLUMN category TYPE character varying,
  ALTER COLUMN status TYPE character varying;

ALTER TABLE public.feedback ALTER COLUMN id DROP IDENTITY IF EXISTS;
ALTER TABLE public.feedback ALTER COLUMN id TYPE bigint USING id::bigint;
ALTER TABLE public.feedback ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;
SELECT setval(pg_get_serial_sequence('public.feedback', 'id'), COALESCE((SELECT max(id) FROM public.feedback), 1), true);

ALTER TABLE public.feedback
  ADD CONSTRAINT feedback_status_check
    CHECK (status::text = ANY (ARRAY['new'::character varying, 'read'::character varying]::text[]));

ALTER TABLE public.page_views ALTER COLUMN id DROP IDENTITY IF EXISTS;
ALTER TABLE public.page_views ALTER COLUMN id TYPE bigint USING id::bigint;
CREATE SEQUENCE IF NOT EXISTS public.page_views_id_seq AS bigint START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.page_views ALTER COLUMN id SET DEFAULT nextval('public.page_views_id_seq'::regclass);
ALTER SEQUENCE public.page_views_id_seq OWNED BY public.page_views.id;
SELECT setval('public.page_views_id_seq', COALESCE((SELECT max(id) FROM public.page_views), 1), true);


-- contributions
ALTER TABLE public.contributions ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.contributions_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.contributions ALTER COLUMN id SET DEFAULT nextval('public.contributions_id_seq'::regclass);
ALTER SEQUENCE public.contributions_id_seq OWNED BY public.contributions.id;
SELECT setval('public.contributions_id_seq', COALESCE((SELECT max(id) FROM public.contributions), 1), true);

-- reports
ALTER TABLE public.reports ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.reports_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.reports ALTER COLUMN id SET DEFAULT nextval('public.reports_id_seq'::regclass);
ALTER SEQUENCE public.reports_id_seq OWNED BY public.reports.id;
SELECT setval('public.reports_id_seq', COALESCE((SELECT max(id) FROM public.reports), 1), true);

-- extra_links
ALTER TABLE public.extra_links ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.extra_links_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.extra_links ALTER COLUMN id SET DEFAULT nextval('public.extra_links_id_seq'::regclass);
ALTER SEQUENCE public.extra_links_id_seq OWNED BY public.extra_links.id;
SELECT setval('public.extra_links_id_seq', COALESCE((SELECT max(id) FROM public.extra_links), 1), true);

-- extra_sections
ALTER TABLE public.extra_sections ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.extra_sections_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.extra_sections ALTER COLUMN id SET DEFAULT nextval('public.extra_sections_id_seq'::regclass);
ALTER SEQUENCE public.extra_sections_id_seq OWNED BY public.extra_sections.id;
SELECT setval('public.extra_sections_id_seq', COALESCE((SELECT max(id) FROM public.extra_sections), 1), true);

-- links
ALTER TABLE public.links ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.links_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.links ALTER COLUMN id SET DEFAULT nextval('public.links_id_seq'::regclass);
ALTER SEQUENCE public.links_id_seq OWNED BY public.links.id;
SELECT setval('public.links_id_seq', COALESCE((SELECT max(id) FROM public.links), 1), true);

-- courses
ALTER TABLE public.courses ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.courses_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.courses ALTER COLUMN id SET DEFAULT nextval('public.courses_id_seq'::regclass);
ALTER SEQUENCE public.courses_id_seq OWNED BY public.courses.id;
SELECT setval('public.courses_id_seq', COALESCE((SELECT max(id) FROM public.courses), 1), true);

-- semesters
ALTER TABLE public.semesters ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.semesters_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.semesters ALTER COLUMN id SET DEFAULT nextval('public.semesters_id_seq'::regclass);
ALTER SEQUENCE public.semesters_id_seq OWNED BY public.semesters.id;
SELECT setval('public.semesters_id_seq', COALESCE((SELECT max(id) FROM public.semesters), 1), true);

-- years
ALTER TABLE public.years ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.years_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.years ALTER COLUMN id SET DEFAULT nextval('public.years_id_seq'::regclass);
ALTER SEQUENCE public.years_id_seq OWNED BY public.years.id;
SELECT setval('public.years_id_seq', COALESCE((SELECT max(id) FROM public.years), 1), true);

-- programs
ALTER TABLE public.programs ALTER COLUMN id DROP IDENTITY IF EXISTS;
CREATE SEQUENCE IF NOT EXISTS public.programs_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER TABLE public.programs ALTER COLUMN id SET DEFAULT nextval('public.programs_id_seq'::regclass);
ALTER SEQUENCE public.programs_id_seq OWNED BY public.programs.id;
SELECT setval('public.programs_id_seq', COALESCE((SELECT max(id) FROM public.programs), 1), true);
