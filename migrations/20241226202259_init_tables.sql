-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tickets
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER     NOT NULL,
    category_id INTEGER     NOT NULL,
    name        VARCHAR(50) NOT NULL,
    description TEXT,
    price       FLOAT       NOT NULL,
    quantity    INTEGER     NOT NULL,
    created_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tickets_tags_associations
(
    id         SERIAL PRIMARY KEY,
    ticket_id  INTEGER   NOT NULL,
    tag_id     INTEGER   NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ticket_id) REFERENCES tickets (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tickets_attachments
(
    id         SERIAL PRIMARY KEY,
    ticket_id  INTEGER   NOT NULL,
    link       VARCHAR   NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ticket_id) REFERENCES tickets (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS responds
(
    id         SERIAL PRIMARY KEY,
    ticket_id  INTEGER   NOT NULL,
    master_id  INTEGER   NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ticket_id) REFERENCES tickets (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tickets_tags_associations;
DROP TABLE IF EXISTS tickets_attachments;
DROP TABLE IF EXISTS responds;
DROP TABLE IF EXISTS tickets;
-- +goose StatementEnd
