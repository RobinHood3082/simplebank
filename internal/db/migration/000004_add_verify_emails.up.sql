CREATE TABLE
    "verify_emails" (
        "id" bigserial PRIMARY KEY,
        "username" varchar NOT NULL,
        "email" varchar NOT NULL,
        "secret_code" varchar NOT NULL,
        "is_used" boolean NOT NULL DEFAULT false,
        "created_at" timestamptz NOT NULL DEFAULT (now ()),
        "expires_at" timestamptz NOT NULL DEFAULT (now () + interval '15 minutes')
    );

CREATE INDEX ON "verify_emails" ("username");

CREATE UNIQUE INDEX ON "verify_emails" ("username", "email");

ALTER TABLE "verify_emails" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");

ALTER TABLE "users"
ADD COLUMN "is_email_verified" boolean NOT NULL DEFAULT false;