-- Search queries and browse-depth steps for admin analytics.
-- Inserts go through the Go API (same pattern as page_views).

CREATE TABLE public.search_events (
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    user_id integer NOT NULL,
    query text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT search_events_pkey PRIMARY KEY (id),
    CONSTRAINT search_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE INDEX search_events_created_at_idx ON public.search_events (created_at DESC);
CREATE INDEX search_events_query_created_at_idx ON public.search_events (query, created_at DESC);

CREATE TABLE public.browse_events (
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    user_id integer NOT NULL,
    step text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT browse_events_pkey PRIMARY KEY (id),
    CONSTRAINT browse_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    CONSTRAINT browse_events_step_chk CHECK (step = ANY (ARRAY['year'::text, 'list'::text]))
);

CREATE INDEX browse_events_created_at_idx ON public.browse_events (created_at DESC);
CREATE INDEX browse_events_step_created_at_idx ON public.browse_events (step, created_at DESC);
