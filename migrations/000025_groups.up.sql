CREATE TABLE groups (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL DEFAULT 'default',
    federation_uuid uuid NOT NULL REFERENCES federations (uuid) ON DELETE CASCADE,
    company_uuid uuid NOT NULL REFERENCES companies (uuid) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX "groups_uuid" ON groups ("uuid");