CREATE TABLE profile_company (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "user_uuid" uuid NOT NULL REFERENCES users (uuid) ON DELETE CASCADE,
    "company_uuid" uuid NOT NULL REFERENCES companies (uuid) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    about text NOT NULL DEFAULT '',
    phones jsonb NOT NULL DEFAULT '[]' :: jsonb,
    emails jsonb NOT NULL DEFAULT '[]' :: jsonb,
    sex int NOT NULL DEFAULT 0,
    birthday date,
    sites jsonb NOT NULL DEFAULT '[]' :: jsonb,
    city jsonb NOT NULL DEFAULT '{}' :: jsonb,
    onboarding date NULL,
    inn varchar(20) NOT NULL DEFAULT '',
    fields jsonb NOT NULL DEFAULT '{}' :: jsonb
);

CREATE UNIQUE INDEX "profile_company_uuid" ON profile_company ("uuid");

CREATE UNIQUE INDEX "profile_company_user_uuid_company_uuid" ON profile_company ("user_uuid", "company_uuid");