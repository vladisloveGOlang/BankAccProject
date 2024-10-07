CREATE TABLE tumblers (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(50) NOT NULL,
    "enable" boolean NOT NULL DEFAULT false,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz
);

CREATE UNIQUE INDEX "tumblers_name" ON tumblers ("name");