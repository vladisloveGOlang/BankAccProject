CREATE TABLE "public"."users" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" varchar(50) NOT NULL DEFAULT '' :: character varying,
    "lname" varchar(50) NOT NULL DEFAULT '' :: character varying,
    "pname" varchar(50) NOT NULL DEFAULT '' :: character varying,
    "email" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "phone" int8 NOT NULL DEFAULT 0,
    "is_valid" bool NOT NULL DEFAULT false,
    "password" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "provider" int8 NOT NULL DEFAULT 0,
    "color" varchar(7) NOT NULL DEFAULT '#000000' :: character varying,
    "has_photo" bool NOT NULL DEFAULT false,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "valid_at" timestamptz,
    "validation_send_at" timestamptz,
    "reset_send_at" timestamptz,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    "position" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    PRIMARY KEY ("uuid")
);

CREATE UNIQUE INDEX "users_email_unique" ON users ("email");

CREATE INDEX "users_created_at" ON users ("created_at" DESC);

CREATE INDEX "users_updated_at" ON users ("updated_at" DESC);