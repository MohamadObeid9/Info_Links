-- Restore the default NO ACTION behaviour on every foreign key touched by
-- 000005, re-adding each one with the same name and target as before.

-- feedback
ALTER TABLE public.feedback DROP CONSTRAINT IF EXISTS feedback_user_id_fkey;
ALTER TABLE public.feedback
  ADD CONSTRAINT feedback_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id);

-- contributions
ALTER TABLE public.contributions DROP CONSTRAINT IF EXISTS contributions_user_id_fkey;
ALTER TABLE public.contributions
  ADD CONSTRAINT contributions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id);

-- reports
ALTER TABLE public.reports DROP CONSTRAINT IF EXISTS reports_user_id_fkey;
ALTER TABLE public.reports
  ADD CONSTRAINT reports_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id);

-- favorite_events
ALTER TABLE public.favorite_events DROP CONSTRAINT IF EXISTS favorite_events_course_id_fkey;
ALTER TABLE public.favorite_events
  ADD CONSTRAINT favorite_events_course_id_fkey
    FOREIGN KEY (course_id) REFERENCES public.courses(id);

ALTER TABLE public.favorite_events DROP CONSTRAINT IF EXISTS favorite_events_user_id_fkey;
ALTER TABLE public.favorite_events
  ADD CONSTRAINT favorite_events_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id);

-- page_views
ALTER TABLE public.page_views DROP CONSTRAINT IF EXISTS page_views_user_id_fkey;
ALTER TABLE public.page_views
  ADD CONSTRAINT page_views_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id);

-- link_clicks (the link_clicks_one_target_chk CHECK from 000002 is left untouched)
ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_user_id_fkey;
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id);

ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_extra_link_id_fkey;
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_extra_link_id_fkey
    FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id);

ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_link_id_fkey;
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_link_id_fkey
    FOREIGN KEY (link_id) REFERENCES public.links(id);

-- extra_links
ALTER TABLE public.extra_links DROP CONSTRAINT IF EXISTS extra_links_section_id_fkey;
ALTER TABLE public.extra_links
  ADD CONSTRAINT extra_links_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES public.extra_sections(id);

-- links
ALTER TABLE public.links DROP CONSTRAINT IF EXISTS links_course_id_fkey;
ALTER TABLE public.links
  ADD CONSTRAINT links_course_id_fkey
    FOREIGN KEY (course_id) REFERENCES public.courses(id);

-- courses
ALTER TABLE public.courses DROP CONSTRAINT IF EXISTS courses_semester_id_fkey;
ALTER TABLE public.courses
  ADD CONSTRAINT courses_semester_id_fkey
    FOREIGN KEY (semester_id) REFERENCES public.semesters(id);

-- semesters
ALTER TABLE public.semesters DROP CONSTRAINT IF EXISTS semesters_year_id_fkey;
ALTER TABLE public.semesters
  ADD CONSTRAINT semesters_year_id_fkey
    FOREIGN KEY (year_id) REFERENCES public.years(id);

-- years
ALTER TABLE public.years DROP CONSTRAINT IF EXISTS years_program_id_fkey;
ALTER TABLE public.years
  ADD CONSTRAINT years_program_id_fkey
    FOREIGN KEY (program_id) REFERENCES public.programs(id);
