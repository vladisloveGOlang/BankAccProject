CREATE TABLE "public"."mails" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "from" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "to" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "subject" varchar(200) NOT NULL DEFAULT '' :: character varying,
    "text" text NOT NULL DEFAULT '' :: text,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    PRIMARY KEY ("uuid")
);