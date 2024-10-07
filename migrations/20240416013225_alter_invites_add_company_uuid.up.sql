ALTER TABLE
    "public"."invites"
ADD
    COLUMN "company_uuid" uuid,
ADD
    FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE CASCADE ON UPDATE CASCADE;