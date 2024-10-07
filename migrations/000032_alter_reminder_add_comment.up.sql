ALTER TABLE
    "reminders"
ADD
    COLUMN "comment" text NOT NULL DEFAULT '' :: character varying;