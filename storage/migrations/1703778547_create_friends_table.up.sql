BEGIN;

create table friends
(
    id        serial not null primary key,
    user_id   uuid   not null,
    friend_id uuid   not null
);

COMMIT;