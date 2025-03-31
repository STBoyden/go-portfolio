-- alter posts table to have published field.
ALTER TABLE "posts"
ADD COLUMN "published" BOOLEAN;

-- set all published column values to false
UPDATE "posts"
SET
    "published" = 'f';

-- require all published fields to be non-null (should be ensured by above)
ALTER TABLE "posts"
ALTER COLUMN "published"
SET
    NOT NULL;

-- set default value for future posts to be FALSE
ALTER TABLE "posts"
ALTER COLUMN "published"
SET DEFAULT FALSE;
