BEGIN;

CREATE TABLE users
(
    id         UUID PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name  VARCHAR(255) NOT NULL,
    birthdate  DATE         NOT NULL,
    biography  VARCHAR(1000),
    city       VARCHAR(255) NOT NULL,
    password   VARCHAR(64) NOT NULL,
    token      VARCHAR(36) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMIT;