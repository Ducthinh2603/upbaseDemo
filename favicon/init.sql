-- EXTENSION
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS favicon (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain_name TEXT NOT NULL,
    image_data BYTEA
);

CREATE INDEX IF NOT EXISTS inx_favicon_domain_name ON favicon (domain_name);

