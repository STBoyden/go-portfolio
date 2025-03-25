-- name: NewAuth :one
INSERT INTO "authorisations" DEFAULT VALUES RETURNING "id";

-- name: GetExpiredAuths :many
SELECT * FROM "authorisations" WHERE "expiry" <= current_timestamp;

-- name: CheckAuthExists :one
SELECT EXISTS (SELECT * FROM "authorisations" WHERE "id" = @id::uuid);

-- name: CheckIfAuthExpired :one
SELECT EXISTS (
  SELECT * FROM "authorisations"
  WHERE "id" = @id::uuid
  AND "expiry" <= current_timestamp
);