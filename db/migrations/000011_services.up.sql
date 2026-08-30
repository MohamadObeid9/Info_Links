CREATE TABLE public.services (
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    title text NOT NULL,
    owner_name text,
    category text,
    emoji text,
    description text,
    logo_url text,
    phone text,
    url text,
    links jsonb DEFAULT '[]'::jsonb NOT NULL,
    status text DEFAULT 'trial' NOT NULL,
    trial boolean DEFAULT true NOT NULL,
    started_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    display_order integer DEFAULT 0 NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT services_pkey PRIMARY KEY (id),
    CONSTRAINT services_status_check CHECK (status IN ('trial', 'active', 'frozen'))
);

CREATE INDEX services_status_expires_at_idx ON public.services (status, expires_at);

CREATE TABLE public.service_clicks (
    id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    service_id integer NOT NULL,
    user_id integer NOT NULL,
    page_context text,
    link_target text,
    clicked_url text,
    clicked_at timestamptz NOT NULL DEFAULT now(),
    device_type text,
    CONSTRAINT service_clicks_pkey PRIMARY KEY (id),
    CONSTRAINT service_clicks_service_id_fkey FOREIGN KEY (service_id) REFERENCES public.services(id) ON DELETE CASCADE,
    CONSTRAINT service_clicks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE INDEX service_clicks_service_id_idx ON public.service_clicks(service_id);
CREATE INDEX service_clicks_clicked_at_idx ON public.service_clicks(clicked_at DESC);
