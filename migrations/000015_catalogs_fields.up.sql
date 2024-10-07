CREATE TABLE "public"."catalog_fields" (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "catalog_uuid" uuid NOT NULL REFERENCES "public"."catalogs"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "hash" varchar(15) NOT NULL,
    "name" varchar(100) NOT NULL,
    "data_type" int8 NOT NULL DEFAULT 0,
    "data_catalog_uuid" uuid REFERENCES "public"."catalogs"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz,
    PRIMARY KEY ("uuid")
);

CREATE UNIQUE INDEX "catalog_fields_unique" ON "public"."catalog_fields"("catalog_uuid", "hash");