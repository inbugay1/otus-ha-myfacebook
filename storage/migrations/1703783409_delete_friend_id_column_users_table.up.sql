BEGIN;

alter table users
    drop column friend_id;

COMMIT;