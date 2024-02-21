SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;
SET default_tablespace = '';
SET default_table_access_method = heap;

CREATE UNLOGGED TABLE client (
    id integer PRIMARY KEY NOT NULL,
    balance integer NOT NULL,
    u_limit integer NOT NULL
);

CREATE UNLOGGED TABLE bank_transaction (
    id SERIAL PRIMARY KEY,
    value integer NOT NULL,
    description varchar(10) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    client_id integer NOT NULL,
    type char(1) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_client_id ON bank_transaction (client_id);
CREATE INDEX IF NOT EXISTS idx_clients_client_id ON client (id);
CREATE INDEX IF NOT EXISTS idx_created_at ON bank_transaction (created_at DESC);

INSERT INTO client (id, balance, u_limit) VALUES (1, 0, 100000);
INSERT INTO client (id, balance, u_limit) VALUES (2, 0, 80000);
INSERT INTO client (id, balance, u_limit) VALUES (3, 0, 1000000);
INSERT INTO client (id, balance, u_limit) VALUES (4, 0, 10000000);
INSERT INTO client (id, balance, u_limit) VALUES (5, 0, 500000);

DROP TYPE IF EXISTS create_transaction_result;
CREATE TYPE create_transaction_result AS ( balance integer, u_limit integer );

CREATE OR REPLACE FUNCTION new_transaction(
    IN client_id integer,
    IN value integer,
    IN description varchar(10),
    IN type char(1)
) RETURNS create_transaction_result AS $$
DECLARE
    clientfound client%rowtype;
    search RECORD;
    ret create_transaction_result;
BEGIN
    SELECT * FROM client INTO clientfound WHERE id = client_id;
    UPDATE client 
        SET balance = CASE 
        WHEN type = 'd' THEN balance - value
            ELSE client.balance + value END 
                WHERE id = client_id AND (type <> 'd' OR (balance - value) >= -u_limit) 
    RETURNING balance, u_limit INTO search;
    IF search.u_limit is NULL THEN
        RAISE EXCEPTION 'limit exceeded';
    ELSE
        INSERT INTO bank_transaction (value, description, created_at, client_id, type) 
        VALUES (value, description, now() at time zone 'utc', client_id, type);
        SELECT search.balance, search.u_limit INTO ret;
    END IF;
    RETURN RET;
END;
$$ LANGUAGE plpgsql;