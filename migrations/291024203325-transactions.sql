CREATE TABLE transactions
(
    uuid               uuid PRIMARY KEY,
    balance_id         INT          NOT NULL,
    delta              float        NOT NULL,
    date               date         NOT NULL,
    tag                VARCHAR(255) NOT NULL
);

CREATE INDEX transactions_balance_id on transactions (balance_id);