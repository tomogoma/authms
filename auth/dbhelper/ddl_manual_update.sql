ALTER TABLE history DROP COLUMN accessMethod;
ALTER TABLE history ADD COLUMN accessMethod STRING NOT NULL DEFAULT 'LOGIN';