-- +migrate Up
CREATE TABLE IF NOT EXISTS workouts.muscle_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT muscle_groups_name_unique UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS workouts.levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT levels_name_unique UNIQUE (name)
);

INSERT INTO workouts.muscle_groups (id, name, created_at, updated_at)
SELECT gen_random_uuid(), trimmed.name, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM (
    SELECT DISTINCT btrim(muscle_group) AS name
    FROM workouts.exercises
    WHERE muscle_group IS NOT NULL AND btrim(muscle_group) <> ''
) AS trimmed
WHERE NOT EXISTS (
    SELECT 1 FROM workouts.muscle_groups mg WHERE mg.name = trimmed.name
);

INSERT INTO workouts.levels (id, name, created_at, updated_at)
SELECT gen_random_uuid(), trimmed.name, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM (
    SELECT DISTINCT btrim(level) AS name
    FROM workouts.exercises
    WHERE level IS NOT NULL AND btrim(level) <> ''
) AS trimmed
WHERE NOT EXISTS (
    SELECT 1 FROM workouts.levels lv WHERE lv.name = trimmed.name
);

ALTER TABLE workouts.exercises
    ADD COLUMN IF NOT EXISTS muscle_group_id UUID,
    ADD COLUMN IF NOT EXISTS level_id UUID;

UPDATE workouts.exercises e
SET muscle_group_id = mg.id
FROM workouts.muscle_groups mg
WHERE e.muscle_group_id IS NULL
  AND e.muscle_group IS NOT NULL
  AND btrim(e.muscle_group) = mg.name;

UPDATE workouts.exercises e
SET level_id = lv.id
FROM workouts.levels lv
WHERE e.level_id IS NULL
  AND e.level IS NOT NULL
  AND btrim(e.level) = lv.name;

ALTER TABLE workouts.exercises
    DROP COLUMN IF EXISTS muscle_group,
    DROP COLUMN IF EXISTS level;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'exercises_muscle_group_id_fkey'
    ) THEN
        ALTER TABLE workouts.exercises
            ADD CONSTRAINT exercises_muscle_group_id_fkey
            FOREIGN KEY (muscle_group_id) REFERENCES workouts.muscle_groups (id);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'exercises_level_id_fkey'
    ) THEN
        ALTER TABLE workouts.exercises
            ADD CONSTRAINT exercises_level_id_fkey
            FOREIGN KEY (level_id) REFERENCES workouts.levels (id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_exercises_muscle_group_id ON workouts.exercises (muscle_group_id);
CREATE INDEX IF NOT EXISTS idx_exercises_level_id ON workouts.exercises (level_id);
