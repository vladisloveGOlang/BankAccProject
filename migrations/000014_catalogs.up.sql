CREATE TABLE "public"."catalogs" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "federation_uuid" uuid NOT NULL,
    "company_uuid" uuid NOT NULL,
    "name" varchar(100) NOT NULL,
    "data_id" int8 NULL DEFAULT 1,
    "description" text NOT NULL DEFAULT '' :: text,
    "created_by" varchar(100) NOT NULL,
    "created_by_uuid" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "field_last_name" int8 NOT NULL DEFAULT 0,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb
);

CREATE UNIQUE INDEX "catalogs_uuid" ON catalogs ("uuid");