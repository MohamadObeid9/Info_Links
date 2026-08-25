-- One catalog row per course code; programs reference it via placements.
-- Links attach to the canonical course so Licence / AISL / IRSM share resources.

CREATE TABLE public.course_placements (
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    course_id integer NOT NULL,
    semester_id integer NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    CONSTRAINT course_placements_pkey PRIMARY KEY (id),
    CONSTRAINT course_placements_course_semester_key UNIQUE (course_id, semester_id),
    CONSTRAINT course_placements_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id) ON DELETE CASCADE,
    CONSTRAINT course_placements_semester_id_fkey FOREIGN KEY (semester_id) REFERENCES public.semesters(id) ON DELETE CASCADE
);

INSERT INTO public.course_placements (course_id, semester_id, display_order)
SELECT id, semester_id, COALESCE(display_order, 0)
FROM public.courses
WHERE semester_id IS NOT NULL;

CREATE TEMP TABLE course_canon AS
SELECT lower(trim(code)) AS code_key,
       (array_agg(id ORDER BY id))[1] AS keep_id,
       array_agg(id ORDER BY id) AS all_ids
FROM public.courses
WHERE code IS NOT NULL AND trim(code) <> ''
GROUP BY 1
HAVING count(*) > 1;

UPDATE public.course_placements pl
SET course_id = c.keep_id
FROM course_canon c
CROSS JOIN unnest(c.all_ids) AS dup_id
WHERE pl.course_id = dup_id AND dup_id <> c.keep_id;

DELETE FROM public.course_placements a
USING public.course_placements b
WHERE a.course_id = b.course_id
  AND a.semester_id = b.semester_id
  AND a.id > b.id;

UPDATE public.links l
SET course_id = c.keep_id
FROM course_canon c
CROSS JOIN unnest(c.all_ids) AS dup_id
WHERE l.course_id = dup_id AND dup_id <> c.keep_id;

UPDATE public.favorite_events fe
SET course_id = c.keep_id
FROM course_canon c
CROSS JOIN unnest(c.all_ids) AS dup_id
WHERE fe.course_id = dup_id AND dup_id <> c.keep_id;

UPDATE public.users u
SET favorite_course_ids = COALESCE((
    SELECT array_agg(DISTINCT COALESCE(mapped.keep_id, x) ORDER BY COALESCE(mapped.keep_id, x))
    FROM unnest(u.favorite_course_ids) AS x
    LEFT JOIN (
        SELECT dup_id, c.keep_id
        FROM course_canon c
        CROSS JOIN unnest(c.all_ids) AS dup_id
        WHERE dup_id <> c.keep_id
    ) mapped ON mapped.dup_id = x
), '{}'::integer[]);

UPDATE public.link_clicks lc
SET link_id = d.keep_id
FROM (
    SELECT min(id) AS keep_id, array_agg(id) AS ids
    FROM public.links
    WHERE course_id IS NOT NULL
    GROUP BY course_id, lower(trim(url))
    HAVING count(*) > 1
) d
CROSS JOIN unnest(d.ids) AS old_id
WHERE lc.link_id = old_id AND old_id <> d.keep_id;

DELETE FROM public.links l
USING (
    SELECT min(id) AS keep_id, course_id, lower(trim(url)) AS url_key
    FROM public.links
    WHERE course_id IS NOT NULL
    GROUP BY course_id, lower(trim(url))
    HAVING count(*) > 1
) d
WHERE l.course_id = d.course_id
  AND lower(trim(l.url)) = d.url_key
  AND l.id <> d.keep_id;

DELETE FROM public.courses co
USING course_canon c
CROSS JOIN unnest(c.all_ids) AS dup_id
WHERE co.id = dup_id AND dup_id <> c.keep_id;

DROP TABLE course_canon;

ALTER TABLE public.courses DROP CONSTRAINT IF EXISTS courses_semester_id_fkey;
ALTER TABLE public.courses DROP COLUMN IF EXISTS semester_id;
ALTER TABLE public.courses DROP COLUMN IF EXISTS display_order;

CREATE UNIQUE INDEX courses_code_lower_uidx
    ON public.courses (lower(trim(code)))
    WHERE code IS NOT NULL AND trim(code) <> '';

CREATE UNIQUE INDEX links_course_url_lower_uidx
    ON public.links (course_id, lower(trim(url)))
    WHERE course_id IS NOT NULL;
