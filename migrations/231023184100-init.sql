CREATE TABLE balance (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    balance FLOAT NOT NULL DEFAULT 0,
    status FLOAT NOT NULL DEFAULT 0,
    day_limit FLOAT NOT NULL DEFAULT 0,
    today_spent FLOAT NOT NULL DEFAULT 0,
    today_added FLOAT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP
);