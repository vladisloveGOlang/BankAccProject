CREATE TABLE "public"."projects" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "description" text NOT NULL DEFAULT '' :: character varying,
    "task_id" int8 NOT NULL DEFAULT 10000,
    "federation_uuid" uuid NOT NULL REFERENCES "public"."federations"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "company_uuid" uuid NOT NULL REFERENCES "public"."companies"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "created_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "responsible_by" varchar(200) NOT NULL DEFAULT '' :: character varying,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "status_graph" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "options" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "status" int8 NOT NULL DEFAULT 0,
    PRIMARY KEY ("uuid")
);

CREATE INDEX "projects_updated_at" ON companies ("updated_at" DESC);

CREATE TABLE "public"."project_users" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "project_uuid" uuid NOT NULL REFERENCES "public"."projects"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "user_uuid" uuid NOT NULL REFERENCES "public"."users"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    CONSTRAINT "fk_project_users_project" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "fk_project_users_user" FOREIGN KEY ("user_uuid") REFERENCES "public"."users"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("uuid")
);

CREATE INDEX "idx_project_users_unique" ON "public"."project_users"("project_uuid", "user_uuid", "deleted_at");

CREATE TABLE "public"."company_fields" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "company_uuid" uuid NOT NULL REFERENCES "public"."companies"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "hash" varchar(15) NOT NULL,
    "name" varchar(100) NOT NULL,
    "icon" varchar(50) NOT NULL DEFAULT '' :: character varying,
    "data_type" int8 NOT NULL DEFAULT 0,
    "required_on_statuses" jsonb NOT NULL DEFAULT '[]' :: jsonb,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    CONSTRAINT "fk_company_fields_company" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("uuid")
);

CREATE UNIQUE INDEX "idx_company_fields_unique" ON "public"."company_fields"("company_uuid", "hash");