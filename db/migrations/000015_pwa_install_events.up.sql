-- PWA install events for student activity timeline / admin analytics.
-- Inserts go through the Go API (same pattern as search_events).

CREATE TABLE public.pwa_install_events (
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    user_id integer NOT NULL,
    device_type text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pwa_install_events_pkey PRIMARY KEY (id),
    CONSTRAINT pwa_install_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
    CONSTRAINT pwa_install_events_device_type_chk CHECK (
        device_type IS NULL OR device_type = ANY (ARRAY['phone'::text, 'laptop'::text])
    )
);

CREATE INDEX pwa_install_events_created_at_idx ON public.pwa_install_events (created_at DESC);
CREATE INDEX pwa_install_events_user_id_created_at_idx ON public.pwa_install_events (user_id, created_at DESC);
