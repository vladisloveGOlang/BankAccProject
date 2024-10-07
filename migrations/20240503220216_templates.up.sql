CREATE TABLE templates (
    "uuid" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "created_by" uuid NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    "federation_uuid" uuid REFERENCES federations (uuid) ON DELETE CASCADE,
    "company_uuid" uuid REFERENCES companies (uuid) ON DELETE CASCADE,
    "project_uuid" uuid REFERENCES projects (uuid) ON DELETE CASCADE,
    "user_uuid" uuid REFERENCES users (uuid) ON DELETE CASCADE,
    "type" varchar(20),
    "template" text,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz
);