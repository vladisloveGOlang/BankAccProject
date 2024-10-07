CREATE TABLE group_users (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    group_uuid uuid NOT NULL REFERENCES groups (uuid) ON DELETE CASCADE,
    user_uuid uuid NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    created_by varchar(100) NOT NULL,
    created_by_uuid uuid NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX "group_user_uuid" ON group_users ("uuid");

CREATE INDEX "group_user_group_uuid_user_uuid" ON group_users ("group_uuid", "user_uuid");

CREATE INDEX "group_user_user_uuid" ON group_users ("user_uuid");