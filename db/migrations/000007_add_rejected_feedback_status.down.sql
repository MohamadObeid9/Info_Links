UPDATE public.feedback SET status = 'read' WHERE status = 'rejected';

ALTER TABLE public.feedback DROP CONSTRAINT IF EXISTS feedback_status_check;
ALTER TABLE public.feedback
  ADD CONSTRAINT feedback_status_check CHECK (status IN ('new', 'read'));
