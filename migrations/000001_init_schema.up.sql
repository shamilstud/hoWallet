-- hoWallet initial schema

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE account_type AS ENUM ('card', 'deposit', 'cash');
CREATE TYPE transaction_type AS ENUM ('income', 'expense', 'transfer');
CREATE TYPE household_role AS ENUM ('owner', 'member');
CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'expired');

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);

-- ============================================================
-- HOUSEHOLDS  (the "wallet group")
-- ============================================================

CREATE TABLE households (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    owner_id   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- HOUSEHOLD MEMBERS  (many-to-many with role)
-- ============================================================

CREATE TABLE household_members (
    household_id UUID NOT NULL REFERENCES households (id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role         household_role NOT NULL DEFAULT 'member',
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (household_id, user_id)
);

CREATE INDEX idx_hm_user ON household_members (user_id);

-- ============================================================
-- INVITATIONS
-- ============================================================

CREATE TABLE invitations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id UUID NOT NULL REFERENCES households (id) ON DELETE CASCADE,
    email        VARCHAR(255) NOT NULL,
    invited_by   UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token        VARCHAR(255) NOT NULL UNIQUE,
    status       invitation_status NOT NULL DEFAULT 'pending',
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invitations_token ON invitations (token);
CREATE INDEX idx_invitations_email ON invitations (email);

-- ============================================================
-- ACCOUNTS  (bank card, deposit, cash, etc.)
-- ============================================================

CREATE TABLE accounts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id UUID NOT NULL REFERENCES households (id) ON DELETE CASCADE,
    name         VARCHAR(255) NOT NULL,
    type         account_type NOT NULL DEFAULT 'card',
    balance      DECIMAL(19, 4) NOT NULL DEFAULT 0,
    currency     VARCHAR(3)  NOT NULL DEFAULT 'UAH',
    created_by   UUID NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_accounts_household ON accounts (household_id);

-- ============================================================
-- TRANSACTIONS
-- ============================================================

CREATE TABLE transactions (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    household_id           UUID NOT NULL REFERENCES households (id) ON DELETE CASCADE,
    type                   transaction_type NOT NULL,
    description            VARCHAR(512) NOT NULL,
    amount                 DECIMAL(19, 4) NOT NULL,
    account_id             UUID NOT NULL REFERENCES accounts (id) ON DELETE CASCADE,
    destination_account_id UUID REFERENCES accounts (id) ON DELETE CASCADE,
    tags                   TEXT[] NOT NULL DEFAULT '{}',
    note                   TEXT,
    transacted_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by             UUID NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_txn_household   ON transactions (household_id);
CREATE INDEX idx_txn_account     ON transactions (account_id);
CREATE INDEX idx_txn_dest        ON transactions (destination_account_id) WHERE destination_account_id IS NOT NULL;
CREATE INDEX idx_txn_transacted  ON transactions (transacted_at);
CREATE INDEX idx_txn_type        ON transactions (type);
CREATE INDEX idx_txn_tags        ON transactions USING GIN (tags);

-- ============================================================
-- REFRESH TOKENS  (stored hashed for security)
-- ============================================================

CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_rt_user ON refresh_tokens (user_id);
CREATE INDEX idx_rt_hash ON refresh_tokens (token_hash);

-- ============================================================
-- HELPER: auto-update updated_at
-- ============================================================

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
