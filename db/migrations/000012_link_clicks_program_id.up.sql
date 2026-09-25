-- Remember which program context a course link was opened from
-- (canonical courses can appear under Licence / AISL / IRSM).
ALTER TABLE public.link_clicks
    ADD COLUMN program_id integer;

ALTER TABLE public.link_clicks
    ADD CONSTRAINT link_clicks_program_id_fkey
    FOREIGN KEY (program_id) REFERENCES public.programs(id) ON DELETE SET NULL;

CREATE INDEX link_clicks_program_id_idx ON public.link_clicks (program_id);
