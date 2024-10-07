CREATE TABLE company_groups_users (
    uuid uuid NOT NULL DEFAULT gen_random_uuid(),
    company_group_uuid uuid NOT NULL REFERENCES company_groups(uuid) ON DELETE CASCADE ON UPDATE CASCADE,
    user_uuid uuid REFERENCES users(uuid) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX "company_groups_users_uuid" ON company_groups_users ("uuid");

CREATE INDEX "company_groups_user_uuid" ON company_groups_users ("user_uuid");