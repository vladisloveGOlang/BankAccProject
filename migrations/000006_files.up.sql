CREATE TABLE "public"."files" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "type" varchar(10) NOT NULL DEFAULT '0' :: smallint,
    "type_uuid" uuid NOT NULL,
    "name" varchar(250) NOT NULL DEFAULT '' :: character varying,
    "object_name" varchar(250) NOT NULL DEFAULT '' :: character varying,
    "size" int8 NOT NULL DEFAULT 0,
    "img_resized" bool NOT NULL DEFAULT false,
    "img_width" int8 NOT NULL DEFAULT 0,
    "img_height" int8 NOT NULL DEFAULT 0,
    "ext" varchar(10) NOT NULL DEFAULT '' :: character varying,
    "mime_type" varchar(50) NOT NULL DEFAULT '' :: character varying,
    "bucket_name" varchar(200) NOT NULL DEFAULT '' :: character varying,
    "endpoint" varchar(30) NOT NULL DEFAULT '' :: character varying,
    "created_by" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "to_deleted_at" timestamptz,
    PRIMARY KEY ("uuid")
);