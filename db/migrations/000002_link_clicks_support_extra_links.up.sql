ALTER TABLE public.link_clicks
  ADD COLUMN extra_link_id bigint;

ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_extra_link_id_fkey
    FOREIGN KEY (extra_link_id) REFERENCES public.extra_links(id);

ALTER TABLE public.link_clicks
  ADD CONSTRAINT link_clicks_one_target_chk CHECK (
    (link_id IS NOT NULL AND extra_link_id IS NULL) OR
    (link_id IS NULL AND extra_link_id IS NOT NULL)
  );
