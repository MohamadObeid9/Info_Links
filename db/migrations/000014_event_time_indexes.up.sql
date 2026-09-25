-- Analytics filters by clicked_at / visited_at alone; existing indexes lead with user_id.
CREATE INDEX link_clicks_clicked_at_idx ON public.link_clicks (clicked_at DESC);
CREATE INDEX page_views_visited_at_idx ON public.page_views (visited_at DESC);
