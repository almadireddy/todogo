create table if not exists users
(
    user_id serial not null
        constraint users_pkey
            primary key,
    username varchar(18) not null
        constraint users_username_key
            unique,
    email varchar(256) not null
        constraint users_email_key
            unique,
    password varchar(256) not null,
    created_at timestamp not null,
    name varchar(256) default 'tom thumb'::character varying not null
);

create table if not exists todoitems
(
    name        varchar(256),
    date        date,
    description text,
    id          serial not null
        constraint todoitems_pk
            primary key,
    user_id     integer
        constraint todoitems_user_id_fkey
            references users
);
