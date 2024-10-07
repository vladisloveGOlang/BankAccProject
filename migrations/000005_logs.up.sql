CREATE TABLE "public"."logs" (
    "uuid" uuid DEFAULT gen_random_uuid(),
    "backend_uuid" varchar(50) DEFAULT 'unknown' :: character varying,
    "header_x_request_id" varchar(255) DEFAULT '' :: character varying,
    "message" text,
    "ip" varchar(255),
    "host" varchar(255),
    "method" varchar(255),
    "request_uri" varchar(255),
    "status" int4,
    "request" jsonb DEFAULT '{}' :: jsonb,
    "agent" varchar(255),
    "referer" varchar(255),
    "start" timestamptz,
    "stop" timestamptz,
    "token" text
);