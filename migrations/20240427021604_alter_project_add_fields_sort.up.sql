ALTER TABLE
    "public"."projects"
ADD
    COLUMN "fields_sort" jsonb NOT NULL DEFAULT '[]',
ALTER COLUMN
    "fields_sort"
SET
    NOT NULL;