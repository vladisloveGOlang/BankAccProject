CREATE SCHEMA IF NOT EXISTS permissions;

CREATE TABLE permissions.groups (
    uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    name character varying(25),
    created_by character varying(255),
    state jsonb NOT NULL DEFAULT '{}',
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE TABLE permissions.users (
    user_uuid uuid NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    created_by character varying(255),
    state jsonb NOT NULL DEFAULT '{}',
    groups uuid [] NOT NULL DEFAULT '{}',
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);