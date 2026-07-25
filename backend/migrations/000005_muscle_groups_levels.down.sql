-- +migrate Down
ALTER TABLE workouts.exercises
    ADD COLUMN IF NOT EXISTS muscle_group VARCHAR(255),
    ADD COLUMN IF NOT EXISTS level VARCHAR(50);

UPDATE workouts.exercises e
SET muscle_group = mg.name
FROM workouts.muscle_groups mg
WHERE e.muscle_group_id = mg.id;

UPDATE workouts.exercises e
SET level = lv.name
FROM workouts.levels lv
WHERE e.level_id = lv.id;

ALTER TABLE workouts.exercises
    DROP CONSTRAINT IF EXISTS exercises_muscle_group_id_fkey,
    DROP CONSTRAINT IF EXISTS exercises_level_id_fkey;

DROP INDEX IF EXISTS workouts.idx_exercises_muscle_group_id;
DROP INDEX IF EXISTS workouts.idx_exercises_level_id;

ALTER TABLE workouts.exercises
    DROP COLUMN IF EXISTS muscle_group_id,
    DROP COLUMN IF EXISTS level_id;

DROP TABLE IF EXISTS workouts.levels;
DROP TABLE IF EXISTS workouts.muscle_groups;
