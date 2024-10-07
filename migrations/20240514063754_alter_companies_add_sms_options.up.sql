ALTER TABLE
    "public"."companies"
ADD
    COLUMN "sms_options" jsonb NOT NULL DEFAULT '{}';