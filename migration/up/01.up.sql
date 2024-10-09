set timezone = 'Europe/Moscow';

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role') THEN
            CREATE TYPE role AS ENUM ('user', 'admin','superAdmin');
        END IF;
END $$;

create table if not exists "user"
(
    id           bigint unique,
    tg_username  text                not null,
    created_at   timestamp           not null,
    channel_from varchar(150)        null,
    user_role    role default 'user' not null,
    primary key (id)
);

create table if not exists question
(
    id int generated always as identity,
    user_id bigint,
    question text,
    is_checked boolean default false,
    primary key (id),
    foreign key (user_id)
        references "user" (id) on delete cascade
);

