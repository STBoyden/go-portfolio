CREATE TABLE IF NOT EXISTS "posts" (
  "id" uuid UNIQUE PRIMARY KEY NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "slug" text UNIQUE NOT NULL,
  "content" json
);
