-- +migrate Up
ALTER TABLE workouts.exercises
    ADD COLUMN IF NOT EXISTS video_urls JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE workouts.exercises
SET video_urls = jsonb_build_array(video_url)
WHERE video_url IS NOT NULL
  AND btrim(video_url) <> ''
  AND (video_urls IS NULL OR video_urls = '[]'::jsonb);

ALTER TABLE workouts.exercises
    DROP COLUMN IF EXISTS video_url;
