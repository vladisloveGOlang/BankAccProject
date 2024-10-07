create extension IF NOT EXISTS ltree;

CREATE TABLE "public"."tasks" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "id" int8 NULL,
    "name" varchar(120) NOT NULL DEFAULT '' :: character varying,
    "icon" varchar(20) NOT NULL DEFAULT '' :: character varying,
    "description" text NOT NULL DEFAULT '' :: text,
    "created_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "managed_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "responsible_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "implement_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "co_workers_by" _text NOT NULL DEFAULT '{}' :: text [],
    "finished_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "watch_by" _text NOT NULL DEFAULT '{}' :: text [],
    "all_people" _text NOT NULL DEFAULT '{}' :: text [],
    "tags" _text NOT NULL DEFAULT '{}' :: text [],
    "federation_uuid" uuid NOT NULL,
    "project_uuid" uuid NOT NULL,
    "status" int8 NOT NULL DEFAULT 0,
    "priority" int8 NOT NULL DEFAULT 10,
    "parent_uuid" uuid,
    "is_epic" bool NOT NULL DEFAULT false,
    "childrens_total" int8 NOT NULL DEFAULT 0,
    "childrens_uuid" _uuid NOT NULL DEFAULT '{}' :: uuid [],
    "comments_total" int8 NOT NULL DEFAULT 0,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "finished_at" timestamptz,
    "finish_to" timestamptz,
    "activity_at" timestamptz NOT NULL DEFAULT now(),
    "duration" int8 NOT NULL DEFAULT 0,
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "fields" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "path" ltree,
    "first_open" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "task_entities" jsonb NOT NULL DEFAULT '[]' :: jsonb,
    "catalog_entities" jsonb NOT NULL DEFAULT '[]' :: jsonb,
    "deal_entities" jsonb NOT NULL DEFAULT '[]' :: jsonb,
    "stops" jsonb NOT NULL DEFAULT '[]' :: jsonb
) partition by hash (project_uuid);

;

CREATE INDEX "tasks_uuid" ON tasks ("uuid");

CREATE INDEX tasks_global ON tasks (
    federation_uuid,
    project_uuid,
    path,
    created_at DESC,
    is_epic,
    status,
    priority,
    all_people
);

create table tasks_p0 partition of tasks for
values
    with (modulus 5, remainder 0);

create table tasks_p1 partition of tasks for
values
    with (modulus 5, remainder 1);

create table tasks_p2 partition of tasks for
values
    with (modulus 5, remainder 2);

create table tasks_p3 partition of tasks for
values
    with (modulus 5, remainder 3);

create table tasks_p4 partition of tasks for
values
    with (modulus 5, remainder 4);