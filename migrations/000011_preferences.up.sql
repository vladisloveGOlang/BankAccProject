CREATE TABLE preferences (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_uuid" uuid NOT NULL REFERENCES users ("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "preferences" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX "preferences_uuid" ON preferences ("uuid");

CREATE UNIQUE INDEX "preferences_user_uuid" ON preferences ("user_uuid");