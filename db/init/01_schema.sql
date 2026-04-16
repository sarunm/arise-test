CREATE TYPE account_status AS ENUM ('ACTIVE', 'INACTIVE', 'FROZEN');
CREATE TYPE transaction_type AS ENUM ('DEPOSIT', 'WITHDRAW', 'TRANSFER');

CREATE TABLE IF NOT EXISTS customers (
    id         BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name  VARCHAR(100) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accounts (
    id          BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    number      VARCHAR(20) NOT NULL UNIQUE,
    balance     BIGINT NOT NULL DEFAULT 0,
    status      account_status NOT NULL DEFAULT 'ACTIVE',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_accounts_customer_id     ON accounts(customer_id);
CREATE INDEX IF NOT EXISTS idx_accounts_customer_status ON accounts(customer_id, status);

CREATE TABLE IF NOT EXISTS transactions (
    id              BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT REFERENCES accounts(id),
    to_account_id   BIGINT REFERENCES accounts(id),
    amount          BIGINT NOT NULL,
    type            transaction_type NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_from ON transactions(from_account_id)
    WHERE from_account_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_to ON transactions(to_account_id)
    WHERE to_account_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
