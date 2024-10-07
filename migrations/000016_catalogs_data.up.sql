CREATE TABLE "public"."catalog_data" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "federation_uuid" uuid NOT NULL,
    "company_uuid" uuid NOT NULL,
    "catalog_uuid" uuid NOT NULL REFERENCES "public"."catalogs"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "id" int8 NULL,
    "description" text NOT NULL DEFAULT '' :: text,
    "created_by" varchar(100) NOT NULL,
    "created_by_uuid" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "fields" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "entities" jsonb NOT NULL DEFAULT '[]' :: jsonb,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb
) partition by hash (company_uuid);

CREATE INDEX "catalog_data_uuid" ON catalog_data ("uuid");

CREATE INDEX catalog_data_global ON catalog_data (company_uuid, created_at DESC);

create table catalog_data_p0 partition of catalog_data for
values
    with (modulus 5, remainder 0);

create table catalog_data_p1 partition of catalog_data for
values
    with (modulus 5, remainder 1);

create table catalog_data_p2 partition of catalog_data for
values
    with (modulus 5, remainder 2);

create table catalog_data_p3 partition of catalog_data for
values
    with (modulus 5, remainder 3);

create table catalog_data_p4 partition of catalog_data for
values
    with (modulus 5, remainder 4);