BEGIN;

CREATE TABLE posts
(
    id         UUID PRIMARY KEY,
    text       TEXT,
    author_id  UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMIT;