CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(50) NOT NULL,
    user_limit INTEGER NOT NULL,
    balance INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS bank_transactions (
    id SERIAL PRIMARY KEY NOT NULL,
    type CHAR(1) NOT NULL,
    description VARCHAR(10) NOT NULL,
    amount INTEGER NOT NULL,
    client_id INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_id ON bank_transactions (client_id);
CREATE INDEX IF NOT EXISTS idx_clients_client_id ON clients (id);

INSERT INTO clients (name, user_limit, balance)
VALUES
    ('David Gilmour', 100000, 0),
    ('Steve Vai', 80000, 0),
    ('Chimbinha', 1000000, 0),
    ('Van Halen', 10000000, 0),
    ('Angus Young', 500000, 0);

CREATE OR REPLACE FUNCTION update_balance()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE clients
    SET balance = CASE
     WHEN NEW.type = 'd' THEN balance - NEW.amount
        ELSE clients.balance + NEW.amount
    END
    WHERE id = NEW.client_id
      AND (NEW.type <> 'd' OR (balance - NEW.amount) >= -user_limit);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER update_balance_trigger
    AFTER INSERT ON bank_transactions
    FOR EACH ROW
EXECUTE FUNCTION update_balance();
