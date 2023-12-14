CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO currencies (id, name) VALUES
    (uuid_generate_v4(), 'UAH'),
    (uuid_generate_v4(), 'EUR'),
    (uuid_generate_v4(), 'USD');