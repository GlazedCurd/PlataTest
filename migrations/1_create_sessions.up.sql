-- TODO: add users
CREATE TYPE quote_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

CREATE TABLE IF NOT EXISTS quotes (
    id serial primary key,
    quote_name TEXT NOT NULL,
    idempotency_key varchar(64) NOT NULL,
    -- FIXED POINT USE NUMBERS FROM 
    -- https://documentation.sas.com/doc/en/fmscdc/5.6/fmspag/p06qd7jikhxltgn1rotrfqopiu4y.htm
    quote NUMERIC(25, 15), 
    status quote_status DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (idempotency_key, quote_name)
);