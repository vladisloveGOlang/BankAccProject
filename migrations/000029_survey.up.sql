CREATE TABLE surveys (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_uuid" uuid NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    "user_email" varchar(255) NOT NULL,
    "name" varchar(255) NOT NULL,
    "body" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE INDEX "surveys_user_uuid" ON surveys ("user_uuid");