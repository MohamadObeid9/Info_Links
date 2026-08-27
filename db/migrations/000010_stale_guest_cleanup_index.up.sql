-- Speeds up periodic DELETE of unclaimed guests that have not been seen recently.
CREATE INDEX users_stale_guests_idx ON public.users (last_seen_at) WHERE is_guest = true;
