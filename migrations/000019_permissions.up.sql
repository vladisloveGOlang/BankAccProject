CREATE TABLE permissions (
    user_uuid uuid NOT NULL REFERENCES users(uuid) ON DELETE CASCADE ON UPDATE CASCADE,
    federation_uuid uuid NOT NULL REFERENCES federations(uuid) ON DELETE CASCADE ON UPDATE CASCADE,
    rules jsonb NOT NULL DEFAULT '{}' :: jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX "permissions_user_uuid" ON permissions ("user_uuid");