ALTER TABLE
    "public"."tasks"
ADD
    COLUMN "company_uuid" uuid NOT NULL DEFAULT gen_random_uuid();