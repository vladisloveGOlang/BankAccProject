CREATE TABLE company_groups (
    uuid uuid NOT NULL DEFAULT gen_random_uuid(),
    federation_uuid uuid NOT NULL REFERENCES federations(uuid) ON DELETE CASCADE ON UPDATE CASCADE,
    company_uuid uuid REFERENCES companies(uuid) ON DELETE CASCADE ON UPDATE CASCADE,
    name varchar(40) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX "company_groups_uuid" ON company_groups ("uuid");

CREATE INDEX "company_groups_company_uuid" ON company_groups ("company_uuid");