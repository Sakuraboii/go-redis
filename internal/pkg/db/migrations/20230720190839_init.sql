-- +goose Up
-- +goose StatementBegin
create table users
(
    id          BIGSERIAL PRIMARY KEY NOT NULL,
    name        varchar,
    second_name varchar,
    surname     varchar
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
