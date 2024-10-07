create extension IF NOT EXISTS ltree;

CREATE TABLE comments (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "created_by" varchar(200) NOT NULL,
    "comment" text NOT NULL DEFAULT '' :: text,
    "reply_uuid" uuid,
    "task_uuid" uuid NOT NULL,
    "people" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "files" _text NOT NULL DEFAULT '{}' :: text [],
    "likes" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "pin" bool NOT NULL DEFAULT false
) partition by hash (task_uuid);

;

CREATE INDEX "comments_uuid" ON comments ("uuid");

CREATE INDEX comments_global ON comments (task_uuid, created_at DESC, people)
where
    people != '{}'
    and deleted_at is null;

create table comments_p0 partition of comments for
values
    with (modulus 5, remainder 0);

create table comments_p1 partition of comments for
values
    with (modulus 5, remainder 1);

create table comments_p2 partition of comments for
values
    with (modulus 5, remainder 2);

create table comments_p3 partition of comments for
values
    with (modulus 5, remainder 3);

create table comments_p4 partition of comments for
values
    with (modulus 5, remainder 4);