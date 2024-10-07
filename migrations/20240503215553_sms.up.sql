CREATE TABLE sms (
    uuid uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    created_by_uuid uuid NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    federation_uuid uuid REFERENCES federations(uuid) ON DELETE CASCADE,
    company_uuid uuid REFERENCES companies(uuid) ON DELETE CASCADE,
    project_uuid uuid REFERENCES projects(uuid) ON DELETE CASCADE,
    text text,
    status character varying(20),
    sent_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    "to" character varying(255),
    "from" character varying(255),
    created_by character varying(255),
    cost integer,
    meta jsonb NOT NULL DEFAULT '{}' :: jsonb
);