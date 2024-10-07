ALTER TABLE
    "public"."projects"
ADD
    COLUMN "status_sort" jsonb NOT NULL DEFAULT '[]';