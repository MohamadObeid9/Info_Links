-- Optional is per offering (placement), not global on the shared course row.
ALTER TABLE public.course_placements
    ADD COLUMN is_optional boolean NOT NULL DEFAULT false;

UPDATE public.course_placements pl
SET is_optional = c.is_optional
FROM public.courses c
WHERE c.id = pl.course_id;
