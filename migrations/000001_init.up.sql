CREATE TABLE balances (
    user_id INT PRIMARY KEY,
    amount NUMERIC(10, 2) NOT NULL CHECK (amount >= 0)
);

CREATE TABLE withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES balances(user_id),
    amount NUMERIC(10, 0) NOT NULL CHECK (amount > 0),
    currency TEXT NOT NULL,
    destination TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',

    UNIQUE(user_id, idempotency_key)
);
