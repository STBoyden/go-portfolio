-- name: NewAuth :one
INSERT INTO "authorisations" (id) VALUES (gen_random_uuid()) RETURNING *;

-- name: GetExpiredAuths :many
SELECT * FROM "authorisations" WHERE "expiry" <= current_timestamp;

-- name: GetAuthByToken :one
SELECT * FROM "authorisations" WHERE "id" = @id::uuid LIMIT 1;

-- name: CheckAuthExists :one
SELECT EXISTS (SELECT * FROM "authorisations" WHERE "id" = @id::uuid);

-- name: CheckIfAuthExpired :one
SELECT EXISTS (
  SELECT * FROM "authorisations"
  WHERE "id" = @id::uuid
  AND "expiry" <= current_timestamp
);
