-- +migrate Down
ALTER TABLE workouts.exercises
    ADD COLUMN IF NOT EXISTS video_url TEXT;

UPDATE workouts.exercises
SET video_url = video_urls ->> 0
WHERE jsonb_typeof(video_urls) = 'array'
  AND jsonb_array_length(video_urls) > 0;

ALTER TABLE workouts.exercises
    DROP COLUMN IF EXISTS video_urls;
