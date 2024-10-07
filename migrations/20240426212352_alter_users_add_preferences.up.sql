ALTER TABLE
    "public"."users"
ADD
    COLUMN "preferences" jsonb NOT NULL DEFAULT '{}';