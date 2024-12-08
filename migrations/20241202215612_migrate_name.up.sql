create table songs (
    id varchar primary key,
    "group" varchar not null,
    song varchar not null,
    release_date date,
    text text,
    link varchar,
    deleted boolean not null default false
);

create unique index songs_idx on songs ("group", song);

