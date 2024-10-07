CREATE TABLE "public"."companies" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "federation_uuid" uuid NOT NULL REFERENCES "public"."federations"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "created_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "created_by_uuid" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "field_last_name" int8 NOT NULL DEFAULT 0,
    CONSTRAINT "fk_federations_companies" FOREIGN KEY ("federation_uuid") REFERENCES "public"."federations"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("uuid")
);

CREATE INDEX "companies_updated_at" ON companies ("updated_at" DESC);

CREATE INDEX "companies_created_at" ON companies ("created_at" DESC);

CREATE TABLE "public"."company_users" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "federation_uuid" uuid NOT NULL REFERENCES "public"."federations"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "company_uuid" uuid NOT NULL REFERENCES "public"."companies"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "user_uuid" uuid NOT NULL,
    "user_groups" jsonb NOT NULL DEFAULT '[]' :: jsonb,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    CONSTRAINT "fk_company_users_federation" FOREIGN KEY ("federation_uuid") REFERENCES "public"."federations"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "fk_company_users_user" FOREIGN KEY ("user_uuid") REFERENCES "public"."users"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "fk_companies_compamy_users" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("uuid")
);

CREATE INDEX "idx_company_users_unique" ON "public"."company_users"("company_uuid", "user_uuid", "deleted_at");

CREATE INDEX "idx_company_users_updated_at" ON "public"."company_users"("updated_at" DESC);