CREATE TABLE public.users (
    id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
    first_name text,
    last_name text,
    number integer,
    is_guest boolean NOT NULL DEFAULT false,
    favorite_course_ids integer[] NOT NULL DEFAULT '{}',
    prefered_lang text NOT NULL DEFAULT 'eng'::text CHECK (prefered_lang = ANY (ARRAY['eng'::text,'fr'::text,'ar'::text])) ,
    prefered_theme text NOT NULL DEFAULT 'system'::text CHECK ( prefered_theme = ANY (ARRAY['system'::text,'dark'::text,'light'::text])),
    created_at timestamp with time zone DEFAULT now(),
    last_seen_at timestamp with time zone DEFAULT now(),
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_number_range_chk CHECK ( number IS NULL OR (number BETWEEN 1 AND 100))
);

CREATE UNIQUE INDEX users_unique_username
    ON public.users (first_name, last_name, number)
    WHERE is_guest = false;

CREATE TABLE public.favorite_events (
    id integer GENERATED ALWAYS AS IDENTITY NOT NULL,
    user_id integer NOT NULL,
    course_id integer NOT NULL,
    action text NOT NULL CHECK (action = ANY (ARRAY['added'::text, 'removed'::text])),
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT favorite_events_pkey PRIMARY KEY (id),
    CONSTRAINT favorite_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id),
    CONSTRAINT favorite_events_course_id_fkey FOREIGN KEY (course_id) REFERENCES public.courses(id)
);

CREATE INDEX favorite_events_user_id_created_at_idx
    ON public.favorite_events (user_id, created_at DESC);

ALTER TABLE public.page_views
    ADD COLUMN user_id integer;

ALTER TABLE public.page_views
    ADD CONSTRAINT page_views_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

CREATE INDEX page_views_user_id_visited_at_idx
    ON public.page_views (user_id, visited_at DESC);

ALTER TABLE public.link_clicks
    ADD COLUMN user_id integer;

ALTER TABLE public.link_clicks
    ADD CONSTRAINT link_clicks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

CREATE INDEX link_clicks_user_id_clicked_at_idx
    ON public.link_clicks (user_id, clicked_at DESC);

ALTER TABLE public.feedback
    ADD COLUMN user_id integer;

ALTER TABLE public.feedback
    ADD CONSTRAINT feedback_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

CREATE INDEX feedback_user_id_created_at_idx
    ON public.feedback (user_id, created_at DESC);

ALTER TABLE public.reports
    ADD COLUMN user_id integer;

ALTER TABLE public.reports
    ADD CONSTRAINT reports_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

CREATE INDEX reports_user_id_created_at_idx
    ON public.reports (user_id, created_at DESC);

ALTER TABLE public.contributions
    ADD COLUMN user_id integer;

ALTER TABLE public.contributions
    ADD CONSTRAINT contributions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

CREATE INDEX contributions_user_id_created_at_idx
    ON public.contributions (user_id, created_at DESC);
