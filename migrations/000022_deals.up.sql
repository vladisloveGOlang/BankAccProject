create extension IF NOT EXISTS ltree;

CREATE TABLE "public"."deals" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "id" int8 NULL,
    "name" varchar(50) NOT NULL DEFAULT '' :: character varying,
    "description" text NOT NULL DEFAULT '' :: text,
    "created_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "tags" text NOT NULL DEFAULT '{}' :: text [],
    "federation_uuid" uuid NOT NULL,
    "company_uuid" uuid NOT NULL,
    "status" int8 NOT NULL DEFAULT 0,
    "priority" int8 NOT NULL DEFAULT 10,
    "finished_at" timestamptz,
    "activity_at" timestamptz NOT NULL DEFAULT now(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "fields" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb
) partition by hash (company_uuid);

CREATE INDEX "deals_uuid" ON deals ("uuid");

CREATE INDEX deals_global ON deals (
    federation_uuid,
    company_uuid,
    status,
    created_at DESC
);

create table deals_p0 partition of deals for
values
    with (modulus 5, remainder 0);

create table deals_p1 partition of deals for
values
    with (modulus 5, remainder 1);

create table deals_p2 partition of deals for
values
    with (modulus 5, remainder 2);

create table deals_p3 partition of deals for
values
    with (modulus 5, remainder 3);

create table deals_p4 partition of deals for
values
    with (modulus 5, remainder 4);