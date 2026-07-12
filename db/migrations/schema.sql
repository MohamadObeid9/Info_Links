-- WARNING: This schema is for context only and is not meant to be run.
-- Table order and constraints may not be valid for execution.

CREATE TABLE public.programs (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  display_order integer DEFAULT 0,
  CONSTRAINT programs_pkey PRIMARY KEY (id)
);
CREATE TABLE public.years (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  program_id integer,
  name text NOT NULL,
  display_order integer DEFAULT 0,
  CONSTRAINT years_pkey PRIMARY KEY (id),
  CONSTRAINT years_program_id_fkey FOREIGN KEY (program_id) REFERENCES public.programs(id)
);
CREATE TABLE public.semesters (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  year_id integer,
  name text NOT NULL,
  display_order integer DEFAULT 0,
  CONSTRAINT semesters_pkey PRIMARY KEY (id),
  CONSTRAINT semesters_year_id_fkey FOREIGN KEY (year_id) REFERENCES public.years(id)
);
CREATE TABLE public.courses (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  semester_id integer,
  name text NOT NULL,
  code text NOT NULL,
  display_order integer DEFAULT 0,
  is_optional boolean DEFAULT false,
  CONSTRAINT courses_pkey PRIMARY KEY (id),
  CONSTRAINT courses_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES public.semesters(id)
);
CREATE TABLE public.links (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
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
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  title text NOT NULL,
  icon text DEFAULT '📁'::text,
  display_order integer DEFAULT 0,
  CONSTRAINT extra_sections_pkey PRIMARY KEY (id)
);
CREATE TABLE public.extra_links (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
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
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  course_name text NOT NULL,
  link_url text DEFAULT ''::text,
  description text NOT NULL,
  status text DEFAULT 'open'::text,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT reports_pkey PRIMARY KEY (id)
);
CREATE TABLE public.contributions (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  course_name text NOT NULL,
  link_url text NOT NULL,
  note text DEFAULT ''::text,
  status text DEFAULT 'pending'::text,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT contributions_pkey PRIMARY KEY (id)
);
CREATE TABLE public.page_views (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  visited_at timestamp with time zone DEFAULT now(),
  page text DEFAULT 'home'::text,
  CONSTRAINT page_views_pkey PRIMARY KEY (id)
);
CREATE TABLE public.feedback (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  category text NOT NULL,
  rating integer NOT NULL CHECK (rating >= 1 AND rating <= 5),
  message text,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  status text NOT NULL DEFAULT 'new'::text CHECK (status = ANY (ARRAY['new'::text, 'read'::text])),
  updated_at timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT feedback_pkey PRIMARY KEY (id)
);
CREATE TABLE public.link_clicks (
  id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
  link_id integer,
  clicked_at timestamp with time zone DEFAULT now(),
  extra_link_id integer,
  CONSTRAINT link_clicks_pkey PRIMARY KEY (id),
  CONSTRAINT link_clicks_link_id_fkey FOREIGN KEY (link_id) REFERENCES public.links(id),
  CONSTRAINT link_clicks_extra_link_id_fkey FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id)
);
