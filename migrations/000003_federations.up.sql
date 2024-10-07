-- Table Definition
CREATE TABLE "public"."federations" (
    "uuid" uuid DEFAULT gen_random_uuid(),
    "id" int8 NULL,
    "name" varchar(50) NOT NULL DEFAULT '' :: character varying,
    "created_by" varchar(100) NOT NULL DEFAULT '' :: character varying,
    "created_by_b_uuid" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    "meta" jsonb NOT NULL DEFAULT '{}' :: jsonb,
    PRIMARY KEY ("uuid")
);

CREATE INDEX "federations_created_at" ON federations ("created_at" DESC);

CREATE INDEX "federations_updated_at" ON federations ("updated_at" DESC);

CREATE TABLE "public"."federation_users" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "federation_uuid" uuid NOT NULL,
    "user_uuid" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    CONSTRAINT "fk_federations_federation_users" FOREIGN KEY ("federation_uuid") REFERENCES "public"."federations"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT "fk_federation_users_user" FOREIGN KEY ("user_uuid") REFERENCES "public"."users"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    PRIMARY KEY ("uuid")
);

CREATE INDEX "idx_federation_users_unique" ON "public"."federation_users"("federation_uuid", "user_uuid", "deleted_at");

CREATE INDEX "idx_federation_users_updated_at" ON "public"."federation_users"("updated_at" DESC);