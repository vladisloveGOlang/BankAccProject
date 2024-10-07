CREATE TABLE reminders (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_by" varchar(200) NOT NULL,
    "created_by_uuid" varchar(200) NOT NULL,
    "task_uuid" uuid NOT NULL,
    "user_uuid" uuid,
    "description" text NOT NULL DEFAULT '' :: text,
    "date_to" timestamptz,
    "date_from" timestamptz,
    "type" varchar(50) NOT NULL DEFAULT '' :: varchar,
    "status" integer NOT NULL DEFAULT 0,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz
);

CREATE INDEX "reminders_uuid" ON reminders ("task_uuid");