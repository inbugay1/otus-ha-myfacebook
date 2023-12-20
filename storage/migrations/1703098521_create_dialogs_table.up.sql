BEGIN;

create table dialogs
(
    id          serial
        primary key,
    sender_id   uuid,
    receiver_id uuid,
    text        varchar(1000),
    created_at  timestamp default CURRENT_TIMESTAMP
);

COMMIT;