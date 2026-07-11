-- WARNING: This schema is for context only and is not meant to be run.
-- Table order and constraints may not be valid for execution.

CREATE TABLE public.programs (
  id integer NOT NULL DEFAULT nextval('programs_id_seq'::regclass),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  display_order integer DEFAULT 0,
  CONSTRAINT programs_pkey PRIMARY KEY (id)
);
CREATE TABLE public.years (
  id integer NOT NULL DEFAULT nextval('years_id_seq'::regclass),
  program_id integer,
  name text NOT NULL,
  display_order integer DEFAULT 0,
  CONSTRAINT years_pkey PRIMARY KEY (id),
  CONSTRAINT years_program_id_fkey FOREIGN KEY (program_id) REFERENCES public.programs(id)
);
CREATE TABLE public.semesters (
  id integer NOT NULL DEFAULT nextval('semesters_id_seq'::regclass),
  year_id integer,
  name text NOT NULL,
  display_order integer DEFAULT 0,
  CONSTRAINT semesters_pkey PRIMARY KEY (id),
  CONSTRAINT semesters_year_id_fkey FOREIGN KEY (year_id) REFERENCES public.years(id)
);
CREATE TABLE public.courses (
  id integer NOT NULL DEFAULT nextval('courses_id_seq'::regclass),
  semester_id integer,
  name text NOT NULL,
  code text NOT NULL,
  display_order integer DEFAULT 0,
  is_optional boolean DEFAULT false,
  CONSTRAINT courses_pkey PRIMARY KEY (id),
  CONSTRAINT courses_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES public.semesters(id)
);
CREATE TABLE public.links (
  id integer NOT NULL DEFAULT nextval('links_id_seq'::regclass),
  course_id integer,
  type text NOT NULL CHECK (type = ANY (ARRAY['telegram'::text, 'drive'::text, 'classroom'::text, 'other'::text])),
  url text NOT NULL,
  label text DEFAULT 'Link'::text,
  note text DEFAULT ''::text,
  display_order integer DEFAULT 0,
  content_type text,
  CONSTRAINT links_pkey PRIMARY KEY (id),
  CONSTRAINT links_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id)
);
CREATE TABLE public.extra_sections (
  id integer NOT NULL DEFAULT nextval('extra_sections_id_seq'::regclass),
  title text NOT NULL,
  icon text DEFAULT '📁'::text,
  display_order integer DEFAULT 0,
  CONSTRAINT extra_sections_pkey PRIMARY KEY (id)
);
CREATE TABLE public.extra_links (
  id integer NOT NULL DEFAULT nextval('extra_links_id_seq'::regclass),
  section_id integer,
  type text NOT NULL CHECK (type = ANY (ARRAY['telegram'::text, 'drive'::text, 'classroom'::text, 'other'::text])),
  url text NOT NULL,
  label text DEFAULT 'Link'::text,
  note text DEFAULT ''::text,
  display_order integer DEFAULT 0,
  content_type text,
  CONSTRAINT extra_links_pkey PRIMARY KEY (id),
  CONSTRAINT extra_links_section_id_fkey FOREIGN KEY (section_id) REFERENCES public.extra_sections(id)
);
CREATE TABLE public.reports (
  id integer NOT NULL DEFAULT nextval('reports_id_seq'::regclass),
  course_name text NOT NULL,
  link_url text DEFAULT ''::text,
  description text NOT NULL,
  status text DEFAULT 'open'::text,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT reports_pkey PRIMARY KEY (id)
);
CREATE TABLE public.contributions (
  id integer NOT NULL DEFAULT nextval('contributions_id_seq'::regclass),
  course_name text NOT NULL,
  link_url text NOT NULL,
  note text DEFAULT ''::text,
  status text DEFAULT 'pending'::text,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT contributions_pkey PRIMARY KEY (id)
);
CREATE TABLE public.page_views (
  id bigint NOT NULL DEFAULT nextval('page_views_id_seq'::regclass),
  visited_at timestamp with time zone DEFAULT now(),
  page text DEFAULT 'home'::text,
  CONSTRAINT page_views_pkey PRIMARY KEY (id)
);
CREATE TABLE public.feedback (
  id bigint GENERATED ALWAYS AS IDENTITY NOT NULL,
  category character varying NOT NULL,
  rating integer NOT NULL CHECK (rating >= 1 AND rating <= 5),
  message text,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  status character varying NOT NULL DEFAULT 'new'::character varying CHECK (status::text = ANY (ARRAY['new'::character varying, 'read'::character varying]::text[])),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT feedback_pkey PRIMARY KEY (id)
);
CREATE TABLE public.link_clicks (
  id bigint GENERATED ALWAYS AS IDENTITY NOT NULL,
  link_id bigint,
  clicked_at timestamp with time zone DEFAULT now(),
  extra_link_id bigint,
  CONSTRAINT link_clicks_pkey PRIMARY KEY (id),
  CONSTRAINT link_clicks_link_id_fkey FOREIGN KEY (link_id) REFERENCES public.links(id),
  CONSTRAINT link_clicks_extra_link_id_fkey FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id)
);
