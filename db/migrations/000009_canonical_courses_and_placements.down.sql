-- Restores semester_id on courses from a single placement. Duplicate
-- program copies that were merged in the up migration are not recreated.

DROP INDEX IF EXISTS public.links_course_url_lower_uidx;
DROP INDEX IF EXISTS public.courses_code_lower_uidx;

ALTER TABLE public.courses ADD COLUMN IF NOT EXISTS semester_id integer;
ALTER TABLE public.courses ADD COLUMN IF NOT EXISTS display_order integer DEFAULT 0;

UPDATE public.courses c
SET semester_id = p.semester_id,
    display_order = p.display_order
FROM (
    SELECT DISTINCT ON (course_id) course_id, semester_id, display_order
    FROM public.course_placements
    ORDER BY course_id, id
) p
WHERE c.id = p.course_id;

ALTER TABLE public.courses
    ADD CONSTRAINT courses_semester_id_fkey
    FOREIGN KEY (semester_id) REFERENCES public.semesters(id) ON DELETE CASCADE;

DROP TABLE IF EXISTS public.course_placements;
