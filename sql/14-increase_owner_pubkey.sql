--
-- Increase the char limit of owner_public_key from 256 to 512.
--

-- pew-pew
\connect blobber_meta;

-- in a transaction
BEGIN;
    ALTER TABLE allocations
        MODIFY COLUMN owner_public_key varchar(512) NOT NULL;
COMMIT;