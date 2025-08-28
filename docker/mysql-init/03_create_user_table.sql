CREATE TABLE IF NOT EXISTS users
(
    id         varchar(100) NOT NULL PRIMARY KEY,
    name       varchar(100) NOT NULL,
    password   varchar(100) NOT NULL,
    role      varchar(100) NULL,
    created_at bigint       NOT NULL,
    updated_at bigint       NOT NULL
);