CREATE TABLE invites (
    "uuid" uuid NOT NULL DEFAULT gen_random_uuid(),
    "federation_uuid" uuid NOT NULL REFERENCES "public"."federations"("uuid") ON DELETE CASCADE ON UPDATE CASCADE,
    "email" varchar(200) NOT NULL,
    "accepted_at" timestamptz,
    "declined_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz
);

CREATE UNIQUE INDEX "invites_uuid" ON invites ("uuid");

CREATE INDEX "invites_email" ON invites ("email");