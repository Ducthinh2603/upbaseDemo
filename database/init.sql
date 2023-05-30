-- EXTENSION
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS favicon (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain_name TEXT NOT NULL,
    image_data BYTEA
);

CREATE INDEX IF NOT EXISTS inx_favicon_domain_name ON favicon (domain_name);

CREATE TABLE IF NOT EXISTS upbase_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS inx_users_email ON upbase_users (email);

CREATE TABLE IF NOT EXISTS upbase_chat_rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP,
    owner_id UUID,
    FOREIGN KEY (owner_id) REFERENCES upbase_users(id)
);

CREATE INDEX IF NOT EXISTS inx_chatroom_onwer_id ON upbase_chat_rooms (owner_id);