CREATE TABLE "public"."project_catalog_data" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "federation_uuid" uuid NOT NULL REFERENCES federations ("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "company_uuid" uuid NOT NULL REFERENCES companies ("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "project_uuid" uuid NOT NULL REFERENCES projects ("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "name" varchar(100) NOT NULL DEFAULT 'reasons',
    "value" varchar(500) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz
) partition by hash (company_uuid);

CREATE INDEX "project_catalog_data_uuid" ON project_catalog_data ("uuid");

create table project_catalog_data_p0 partition of project_catalog_data for
values
    with (modulus 5, remainder 0);

create table project_catalog_data_p1 partition of project_catalog_data for
values
    with (modulus 5, remainder 1);

create table project_catalog_data_p2 partition of project_catalog_data for
values
    with (modulus 5, remainder 2);

create table project_catalog_data_p3 partition of project_catalog_data for
values
    with (modulus 5, remainder 3);

create table project_catalog_data_p4 partition of project_catalog_data for
values
    with (modulus 5, remainder 4);