--
-- PostgreSQL database dump
--

\restrict ARubgEUTEUEvYQmCoxvtf2WBFdVHwFLCBVtf7qtn09SEwpJuYJgJrLTtPvY8Re6

-- Dumped from database version 17.6
-- Dumped by pg_dump version 18.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: rls_auto_enable(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.rls_auto_enable() RETURNS event_trigger
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
DECLARE
  cmd record;
BEGIN
  FOR cmd IN
    SELECT *
    FROM pg_event_trigger_ddl_commands()
    WHERE command_tag IN ('CREATE TABLE', 'CREATE TABLE AS', 'SELECT INTO')
      AND object_type IN ('table','partitioned table')
  LOOP
     IF cmd.schema_name IS NOT NULL AND cmd.schema_name IN ('public') AND cmd.schema_name NOT IN ('pg_catalog','information_schema') AND cmd.schema_name NOT LIKE 'pg_toast%' AND cmd.schema_name NOT LIKE 'pg_temp%' THEN
      BEGIN
        EXECUTE format('alter table if exists %s enable row level security', cmd.object_identity);
        RAISE LOG 'rls_auto_enable: enabled RLS on %', cmd.object_identity;
      EXCEPTION
        WHEN OTHERS THEN
          RAISE LOG 'rls_auto_enable: failed to enable RLS on %', cmd.object_identity;
      END;
     ELSE
        RAISE LOG 'rls_auto_enable: skip % (either system schema or not in enforced list: %.)', cmd.object_identity, cmd.schema_name;
     END IF;
  END LOOP;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: contributions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contributions (
    id integer NOT NULL,
    course_name text NOT NULL,
    link_url text NOT NULL,
    note text DEFAULT ''::text,
    status text DEFAULT 'pending'::text,
    created_at timestamp with time zone DEFAULT now(),
    user_id integer
);


--
-- Name: contributions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.contributions ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.contributions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: courses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.courses (
    id integer NOT NULL,
    semester_id integer,
    name text NOT NULL,
    code text NOT NULL,
    display_order integer DEFAULT 0,
    is_optional boolean DEFAULT false
);


--
-- Name: courses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.courses ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.courses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: extra_links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.extra_links (
    id integer NOT NULL,
    section_id integer,
    type text NOT NULL,
    url text NOT NULL,
    label text DEFAULT 'Link'::text,
    note text DEFAULT ''::text,
    display_order integer DEFAULT 0,
    content_type text,
    CONSTRAINT extra_links_type_check CHECK ((type = ANY (ARRAY['telegram'::text, 'drive'::text, 'classroom'::text, 'other'::text])))
);


--
-- Name: extra_links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.extra_links ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.extra_links_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: extra_sections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.extra_sections (
    id integer NOT NULL,
    title text NOT NULL,
    icon text DEFAULT '📁'::text,
    display_order integer DEFAULT 0
);


--
-- Name: extra_sections_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.extra_sections ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.extra_sections_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: favorite_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.favorite_events (
    id integer NOT NULL,
    user_id integer NOT NULL,
    course_id integer NOT NULL,
    action text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT favorite_events_action_check CHECK ((action = ANY (ARRAY['added'::text, 'removed'::text])))
);


--
-- Name: favorite_events_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.favorite_events ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.favorite_events_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: feedback; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.feedback (
    id integer NOT NULL,
    category text NOT NULL,
    rating integer NOT NULL,
    message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    status text DEFAULT 'new'::text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    user_id integer,
    CONSTRAINT feedback_rating_check CHECK (((rating >= 1) AND (rating <= 5))),
    CONSTRAINT feedback_status_check CHECK ((status = ANY (ARRAY['new'::text, 'read'::text, 'rejected'::text])))
);


--
-- Name: feedback_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.feedback ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.feedback_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: link_clicks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.link_clicks (
    id integer NOT NULL,
    link_id integer,
    clicked_at timestamp with time zone DEFAULT now(),
    extra_link_id integer,
    user_id integer,
    CONSTRAINT link_clicks_one_target_chk CHECK ((((link_id IS NOT NULL) AND (extra_link_id IS NULL)) OR ((link_id IS NULL) AND (extra_link_id IS NOT NULL))))
);


--
-- Name: link_clicks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.link_clicks ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.link_clicks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: links; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.links (
    id integer NOT NULL,
    course_id integer,
    type text NOT NULL,
    url text NOT NULL,
    label text DEFAULT 'Link'::text,
    note text DEFAULT ''::text,
    display_order integer DEFAULT 0,
    content_type text,
    CONSTRAINT links_type_check CHECK ((type = ANY (ARRAY['telegram'::text, 'drive'::text, 'classroom'::text, 'other'::text])))
);


--
-- Name: links_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.links ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.links_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: page_views; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.page_views (
    id integer NOT NULL,
    visited_at timestamp with time zone DEFAULT now(),
    page text DEFAULT 'home'::text,
    user_id integer,
    device_type text,
    CONSTRAINT page_views_device_type_chk CHECK (((device_type IS NULL) OR (device_type = ANY (ARRAY['phone'::text, 'laptop'::text]))))
);


--
-- Name: page_views_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.page_views ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.page_views_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: programs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.programs (
    id integer NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    display_order integer DEFAULT 0
);


--
-- Name: programs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.programs ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.programs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reports (
    id integer NOT NULL,
    course_name text NOT NULL,
    link_url text DEFAULT ''::text,
    description text NOT NULL,
    status text DEFAULT 'open'::text,
    created_at timestamp with time zone DEFAULT now(),
    user_id integer
);


--
-- Name: reports_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.reports ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.reports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: semesters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.semesters (
    id integer NOT NULL,
    year_id integer,
    name text NOT NULL,
    display_order integer DEFAULT 0
);


--
-- Name: semesters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.semesters ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.semesters_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    first_name text,
    last_name text,
    number integer,
    is_guest boolean DEFAULT false NOT NULL,
    favorite_course_ids integer[] DEFAULT '{}'::integer[] NOT NULL,
    prefered_lang text DEFAULT 'eng'::text NOT NULL,
    prefered_theme text DEFAULT 'system'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    last_seen_at timestamp with time zone DEFAULT now(),
    CONSTRAINT users_number_range_chk CHECK (((number IS NULL) OR ((number >= 1) AND (number <= 100)))),
    CONSTRAINT users_prefered_lang_check CHECK ((prefered_lang = ANY (ARRAY['eng'::text, 'fr'::text, 'ar'::text]))),
    CONSTRAINT users_prefered_theme_check CHECK ((prefered_theme = ANY (ARRAY['system'::text, 'dark'::text, 'light'::text])))
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: years; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.years (
    id integer NOT NULL,
    program_id integer,
    name text NOT NULL,
    display_order integer DEFAULT 0
);


--
-- Name: years_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.years ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.years_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: contributions contributions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contributions
    ADD CONSTRAINT contributions_pkey PRIMARY KEY (id);


--
-- Name: courses courses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_pkey PRIMARY KEY (id);


--
-- Name: extra_links extra_links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.extra_links
    ADD CONSTRAINT extra_links_pkey PRIMARY KEY (id);


--
-- Name: extra_sections extra_sections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.extra_sections
    ADD CONSTRAINT extra_sections_pkey PRIMARY KEY (id);


--
-- Name: favorite_events favorite_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_events
    ADD CONSTRAINT favorite_events_pkey PRIMARY KEY (id);


--
-- Name: feedback feedback_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feedback
    ADD CONSTRAINT feedback_pkey PRIMARY KEY (id);


--
-- Name: link_clicks link_clicks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.link_clicks
    ADD CONSTRAINT link_clicks_pkey PRIMARY KEY (id);


--
-- Name: links links_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.links
    ADD CONSTRAINT links_pkey PRIMARY KEY (id);


--
-- Name: page_views page_views_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_views
    ADD CONSTRAINT page_views_pkey PRIMARY KEY (id);


--
-- Name: programs programs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.programs
    ADD CONSTRAINT programs_pkey PRIMARY KEY (id);


--
-- Name: programs programs_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.programs
    ADD CONSTRAINT programs_slug_key UNIQUE (slug);


--
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: semesters semesters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.semesters
    ADD CONSTRAINT semesters_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: years years_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.years
    ADD CONSTRAINT years_pkey PRIMARY KEY (id);


--
-- Name: contributions_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX contributions_user_id_created_at_idx ON public.contributions USING btree (user_id, created_at DESC);


--
-- Name: favorite_events_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX favorite_events_user_id_created_at_idx ON public.favorite_events USING btree (user_id, created_at DESC);


--
-- Name: feedback_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX feedback_user_id_created_at_idx ON public.feedback USING btree (user_id, created_at DESC);


--
-- Name: idx_feedback_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_feedback_created_at ON public.feedback USING btree (created_at DESC);


--
-- Name: idx_feedback_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_feedback_status ON public.feedback USING btree (status);


--
-- Name: link_clicks_user_id_clicked_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX link_clicks_user_id_clicked_at_idx ON public.link_clicks USING btree (user_id, clicked_at DESC);


--
-- Name: page_views_user_id_visited_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX page_views_user_id_visited_at_idx ON public.page_views USING btree (user_id, visited_at DESC);


--
-- Name: reports_user_id_created_at_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX reports_user_id_created_at_idx ON public.reports USING btree (user_id, created_at DESC);


--
-- Name: users_unique_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX users_unique_username ON public.users USING btree (first_name, last_name, number) WHERE (is_guest = false);


--
-- Name: contributions contributions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contributions
    ADD CONSTRAINT contributions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: courses courses_semester_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.courses
    ADD CONSTRAINT courses_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES public.semesters(id) ON DELETE CASCADE;


--
-- Name: extra_links extra_links_section_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.extra_links
    ADD CONSTRAINT extra_links_section_id_fkey FOREIGN KEY (section_id) REFERENCES public.extra_sections(id) ON DELETE CASCADE;


--
-- Name: favorite_events favorite_events_course_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_events
    ADD CONSTRAINT favorite_events_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE;


--
-- Name: favorite_events favorite_events_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorite_events
    ADD CONSTRAINT favorite_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: feedback feedback_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.feedback
    ADD CONSTRAINT feedback_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: link_clicks link_clicks_extra_link_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.link_clicks
    ADD CONSTRAINT link_clicks_extra_link_id_fkey FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id) ON DELETE CASCADE;


--
-- Name: link_clicks link_clicks_link_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.link_clicks
    ADD CONSTRAINT link_clicks_link_id_fkey FOREIGN KEY (link_id) REFERENCES public.links(id) ON DELETE CASCADE;


--
-- Name: link_clicks link_clicks_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.link_clicks
    ADD CONSTRAINT link_clicks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: links links_course_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.links
    ADD CONSTRAINT links_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE;


--
-- Name: page_views page_views_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_views
    ADD CONSTRAINT page_views_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reports reports_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: semesters semesters_year_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.semesters
    ADD CONSTRAINT semesters_year_id_fkey FOREIGN KEY (year_id) REFERENCES public.years(id) ON DELETE CASCADE;


--
-- Name: years years_program_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.years
    ADD CONSTRAINT years_program_id_fkey FOREIGN KEY (program_id) REFERENCES public.programs(id) ON DELETE CASCADE;


--
-- Name: feedback Allow anon insert feedback; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow anon insert feedback" ON public.feedback FOR INSERT TO anon WITH CHECK (true);


--
-- Name: feedback Allow anonymous insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow anonymous insert" ON public.feedback FOR INSERT TO anon WITH CHECK (true);


--
-- Name: feedback Allow authenticated manage; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow authenticated manage" ON public.feedback TO authenticated USING (true);


--
-- Name: feedback Allow authenticated users to delete feedback; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow authenticated users to delete feedback" ON public.feedback FOR DELETE USING ((auth.role() = 'authenticated'::text));


--
-- Name: feedback Allow authenticated users to read feedback; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow authenticated users to read feedback" ON public.feedback FOR SELECT TO authenticated, anon USING ((auth.role() = 'authenticated'::text));


--
-- Name: feedback Allow authenticated users to update feedback; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow authenticated users to update feedback" ON public.feedback FOR UPDATE USING ((auth.role() = 'authenticated'::text)) WITH CHECK ((auth.role() = 'authenticated'::text));


--
-- Name: courses Allow public read courses; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow public read courses" ON public.courses FOR SELECT USING (true);


--
-- Name: links Allow public read links; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow public read links" ON public.links FOR SELECT USING (true);


--
-- Name: programs Allow public read programs; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "Allow public read programs" ON public.programs FOR SELECT USING (true);


--
-- Name: contributions admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.contributions USING ((auth.role() = 'authenticated'::text));


--
-- Name: courses admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.courses USING ((auth.role() = 'authenticated'::text));


--
-- Name: extra_links admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.extra_links USING ((auth.role() = 'authenticated'::text));


--
-- Name: extra_sections admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.extra_sections USING ((auth.role() = 'authenticated'::text));


--
-- Name: links admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.links USING ((auth.role() = 'authenticated'::text));


--
-- Name: programs admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.programs USING ((auth.role() = 'authenticated'::text));


--
-- Name: reports admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.reports USING ((auth.role() = 'authenticated'::text));


--
-- Name: semesters admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.semesters USING ((auth.role() = 'authenticated'::text));


--
-- Name: years admin all; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin all" ON public.years USING ((auth.role() = 'authenticated'::text));


--
-- Name: page_views admin read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "admin read" ON public.page_views FOR SELECT USING ((auth.role() = 'authenticated'::text));


--
-- Name: link_clicks anon_insert_link_clicks; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY anon_insert_link_clicks ON public.link_clicks FOR INSERT TO anon WITH CHECK (true);


--
-- Name: link_clicks auth_select_link_clicks; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY auth_select_link_clicks ON public.link_clicks FOR SELECT TO authenticated USING (true);


--
-- Name: contributions; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.contributions ENABLE ROW LEVEL SECURITY;

--
-- Name: contributions contributions_anon_insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY contributions_anon_insert ON public.contributions FOR INSERT TO anon WITH CHECK (true);


--
-- Name: contributions contributions_auth_delete; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY contributions_auth_delete ON public.contributions FOR DELETE TO authenticated USING (true);


--
-- Name: contributions contributions_auth_select; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY contributions_auth_select ON public.contributions FOR SELECT TO authenticated USING (true);


--
-- Name: contributions contributions_auth_update; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY contributions_auth_update ON public.contributions FOR UPDATE TO authenticated USING (true) WITH CHECK (true);


--
-- Name: courses; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.courses ENABLE ROW LEVEL SECURITY;

--
-- Name: extra_links; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.extra_links ENABLE ROW LEVEL SECURITY;

--
-- Name: extra_sections; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.extra_sections ENABLE ROW LEVEL SECURITY;

--
-- Name: favorite_events; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.favorite_events ENABLE ROW LEVEL SECURITY;

--
-- Name: feedback; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.feedback ENABLE ROW LEVEL SECURITY;

--
-- Name: link_clicks; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.link_clicks ENABLE ROW LEVEL SECURITY;

--
-- Name: links; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.links ENABLE ROW LEVEL SECURITY;

--
-- Name: page_views; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.page_views ENABLE ROW LEVEL SECURITY;

--
-- Name: page_views page_views_anon_insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY page_views_anon_insert ON public.page_views FOR INSERT TO anon WITH CHECK (true);


--
-- Name: page_views page_views_auth_select; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY page_views_auth_select ON public.page_views FOR SELECT TO authenticated USING (true);


--
-- Name: programs; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.programs ENABLE ROW LEVEL SECURITY;

--
-- Name: contributions public insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public insert" ON public.contributions FOR INSERT WITH CHECK (true);


--
-- Name: page_views public insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public insert" ON public.page_views FOR INSERT WITH CHECK (true);


--
-- Name: reports public insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public insert" ON public.reports FOR INSERT WITH CHECK (true);


--
-- Name: contributions public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.contributions FOR SELECT USING (true);


--
-- Name: courses public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.courses FOR SELECT USING (true);


--
-- Name: extra_links public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.extra_links FOR SELECT USING (true);


--
-- Name: extra_sections public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.extra_sections FOR SELECT USING (true);


--
-- Name: links public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.links FOR SELECT USING (true);


--
-- Name: programs public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.programs FOR SELECT USING (true);


--
-- Name: reports public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.reports FOR SELECT USING (true);


--
-- Name: semesters public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.semesters FOR SELECT USING (true);


--
-- Name: years public read; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY "public read" ON public.years FOR SELECT USING (true);


--
-- Name: reports; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.reports ENABLE ROW LEVEL SECURITY;

--
-- Name: reports reports_anon_insert; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY reports_anon_insert ON public.reports FOR INSERT TO anon WITH CHECK (true);


--
-- Name: reports reports_auth_delete; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY reports_auth_delete ON public.reports FOR DELETE TO authenticated USING (true);


--
-- Name: reports reports_auth_select; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY reports_auth_select ON public.reports FOR SELECT TO authenticated USING (true);


--
-- Name: reports reports_auth_update; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY reports_auth_update ON public.reports FOR UPDATE TO authenticated USING (true) WITH CHECK (true);


--
-- Name: schema_migrations; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.schema_migrations ENABLE ROW LEVEL SECURITY;

--
-- Name: semesters; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.semesters ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

--
-- Name: years; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.years ENABLE ROW LEVEL SECURITY;

--
-- PostgreSQL database dump complete
--

\unrestrict ARubgEUTEUEvYQmCoxvtf2WBFdVHwFLCBVtf7qtn09SEwpJuYJgJrLTtPvY8Re6
