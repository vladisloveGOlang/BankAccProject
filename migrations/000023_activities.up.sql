CREATE TABLE "public"."activities" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_by" varchar(250) NOT NULL,
    "created_by_uuid" uuid NOT NULL,
    "entity_type" varchar(20) NOT NULL,
    "entity_uuid" uuid NOT NULL,
    "description" varchar(250) NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "entities" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb
) partition by hash (entity_uuid);

CREATE INDEX "activities_uuid" ON activities ("uuid");

CREATE INDEX "activities_entity_uuid" ON activities ("entity_uuid");

create table activities_p0 partition of activities for
values
    with (modulus 5, remainder 0);

create table activities_p1 partition of activities for
values
    with (modulus 5, remainder 1);

create table activities_p2 partition of activities for
values
    with (modulus 5, remainder 2);

create table activities_p3 partition of activities for
values
    with (modulus 5, remainder 3);

create table activities_p4 partition of activities for
values
    with (modulus 5, remainder 4);