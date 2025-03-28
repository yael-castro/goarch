DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age SMALLINT NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    -- Common fields
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP DEFAULT NULL
);

DROP TABLE IF EXISTS outbox_messages;
CREATE TABLE outbox_messages (
    id SERIAL PRIMARY KEY,
    topic VARCHAR NOT NULL,
    idempotency_key BYTEA UNIQUE,
    partition_key BYTEA,
    headers BYTEA,
    value BYTEA NOT NULL,
    delivered_at TIMESTAMP DEFAULT NULL,
    -- Common fields
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP DEFAULT NULL
);