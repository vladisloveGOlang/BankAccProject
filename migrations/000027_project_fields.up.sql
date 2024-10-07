CREATE TABLE project_fields (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "company_uuid" uuid NOT NULL REFERENCES companies (uuid) ON DELETE CASCADE,
    "project_uuid" uuid NOT NULL REFERENCES projects (uuid) ON DELETE CASCADE,
    "company_field_uuid" uuid NOT NULL REFERENCES company_fields (uuid) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    required_on_statuses jsonb NOT NULL DEFAULT '[]' :: jsonb,
    deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX "project_fields_uuid" ON project_fields ("uuid");

CREATE INDEX "project_fields_company_uuid" ON project_fields ("company_uuid");