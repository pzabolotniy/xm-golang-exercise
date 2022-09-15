-- +migrate Up
-- +migrate StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS companies (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    code integer NOT NULL,
    country varchar(512) NOT NULL,
    website varchar(2048) NOT NULL, -- max URL length is 2048
    phone varchar(64) NOT NULL, -- phone number max length - 15 digits + whitespace and dashes + extensions
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TABLE IF EXISTS companies;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +migrate StatementEnd