-- Add referential actions to the existing foreign keys.
-- Until now every FK used the default NO ACTION, so deleting a program, course,
-- link or user was either blocked outright or left rows pointing at a parent the
-- application no longer shows.
--
-- Content hierarchy and per-user analytics rows are meaningless without their
-- parent, so they cascade. The three user-submission tables (reports,
-- contributions, feedback) instead fall back to SET NULL: those columns are
-- already nullable for legacy anonymous rows, and deleting a student must not
-- erase submissions an admin still has to act on.
--
-- Postgres cannot alter a foreign key in place, so each change is a drop
-- followed by a re-add.

-- years
ALTER TABLE public.years DROP CONSTRAINT IF EXISTS years_program_id_fkey;
ALTER TABLE public.years
  ADD CONSTRAINT years_program_id_fkey
    FOREIGN KEY (program_id) REFERENCES public.programs(id) ON DELETE CASCADE;

-- semesters
ALTER TABLE public.semesters DROP CONSTRAINT IF EXISTS semesters_year_id_fkey;
ALTER TABLE public.semesters
  ADD CONSTRAINT semesters_year_id_fkey
    FOREIGN KEY (year_id) REFERENCES public.years(id) ON DELETE CASCADE;

-- courses
ALTER TABLE public.courses DROP CONSTRAINT IF EXISTS courses_semester_id_fkey;
ALTER TABLE public.courses
  ADD CONSTRAINT courses_semester_id_fkey
    FOREIGN KEY (semester_id) REFERENCES public.semesters(id) ON DELETE CASCADE;

-- links
ALTER TABLE public.links DROP CONSTRAINT IF EXISTS links_course_id_fkey;
ALTER TABLE public.links
  ADD CONSTRAINT links_course_id_fkey
    FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE;

-- extra_links
ALTER TABLE public.extra_links DROP CONSTRAINT IF EXISTS extra_links_section_id_fkey;
ALTER TABLE public.extra_links
  ADD CONSTRAINT extra_links_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES public.extra_sections(id) ON DELETE CASCADE;

-- link_clicks (the link_clicks_one_target_chk CHECK from 000002 is left untouched)
ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_link_id_fkey;
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_link_id_fkey
    FOREIGN KEY (link_id) REFERENCES public.links(id) ON DELETE CASCADE;

ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_extra_link_id_fkey;
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_extra_link_id_fkey
    FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id) ON DELETE CASCADE;

ALTER TABLE public.link_clicks DROP CONSTRAINT IF EXISTS link_clicks_user_id_fkey;
ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- page_views
ALTER TABLE public.page_views DROP CONSTRAINT IF EXISTS page_views_user_id_fkey;
ALTER TABLE public.page_views
  ADD CONSTRAINT page_views_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

-- favorite_events
ALTER TABLE public.favorite_events DROP CONSTRAINT IF EXISTS favorite_events_user_id_fkey;
ALTER TABLE public.favorite_events
  ADD CONSTRAINT favorite_events_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

ALTER TABLE public.favorite_events DROP CONSTRAINT IF EXISTS favorite_events_course_id_fkey;
ALTER TABLE public.favorite_events
  ADD CONSTRAINT favorite_events_course_id_fkey
    FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE;

-- reports: keep the submission, drop the attribution
ALTER TABLE public.reports DROP CONSTRAINT IF EXISTS reports_user_id_fkey;
ALTER TABLE public.reports
  ADD CONSTRAINT reports_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

-- contributions: keep the submission, drop the attribution
ALTER TABLE public.contributions DROP CONSTRAINT IF EXISTS contributions_user_id_fkey;
ALTER TABLE public.contributions
  ADD CONSTRAINT contributions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;

-- feedback: keep the submission, drop the attribution
ALTER TABLE public.feedback DROP CONSTRAINT IF EXISTS feedback_user_id_fkey;
ALTER TABLE public.feedback
  ADD CONSTRAINT feedback_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;
