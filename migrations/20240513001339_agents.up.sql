CREATE TABLE agents (
    uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    created_by_uuid uuid NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    created_by varchar(255) NOT NULL,
    federation_uuid uuid REFERENCES federations(uuid) ON DELETE CASCADE,
    company_uuid uuid REFERENCES companies(uuid) ON DELETE CASCADE,
    project_uuid uuid REFERENCES projects(uuid) ON DELETE CASCADE,
    name varchar(255) NOT NULL,
    contacts jsonb NOT NULL DEFAULT '{}' :: jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    meta jsonb NOT NULL DEFAULT '{}' :: jsonb
);