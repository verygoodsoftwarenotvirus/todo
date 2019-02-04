CREATE TABLE IF NOT EXISTS items (
    "id" BIGSERIAL NOT NULL PRIMARY KEY,
    "name" text NOT NULL,
    "details" TEXT NOT NULL DEFAULT '',
    "created_on" timestamp NOT NULL DEFAULT NOW(),
    "updated_on" timestamp,
    "archived_on" timestamp,
    "belongs_to" bigint NOT NULL,
    FOREIGN KEY ("belongs_to") REFERENCES "users"("id")
);
