CREATE TABLE company_tags (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "federation_uuid" uuid NOT NULL REFERENCES federations (uuid) ON DELETE CASCADE,
    "company_uuid" uuid NOT NULL REFERENCES companies (uuid) ON DELETE CASCADE,
    "name" varchar(100) NOT NULL,
    "color" varchar(10) NOT NULL,
    created_by varchar(100) NOT NULL DEFAULT '',
    created_by_uuid uuid NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE INDEX "company_tags_company_uuid" ON company_tags ("company_uuid");