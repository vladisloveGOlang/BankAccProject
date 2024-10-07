CREATE TABLE company_priorities (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "company_uuid" uuid NOT NULL REFERENCES companies (uuid) ON DELETE CASCADE,
    "number" integer NOT NULL DEFAULT 10,
    "name" varchar(255) NOT NULL,
    "color" varchar(255) NOT NULL DEFAULT '#000000',
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);