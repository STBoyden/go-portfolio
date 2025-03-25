CREATE TABLE IF NOT EXISTS "authorisations" (
  "id" uuid UNIQUE PRIMARY KEY NOT NULL,
  "expiry" timestamp NOT NULL DEFAULT (now() + interval '2 hour')
);
