BEGIN;

ALTER TABLE users
    ADD friend_id VARCHAR(36) NOT NULL DEFAULT '';

COMMIT;