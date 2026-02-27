-- Rollback initial schema

DROP TRIGGER IF EXISTS trg_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS trg_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS household_members;
DROP TABLE IF EXISTS households;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS invitation_status;
DROP TYPE IF EXISTS household_role;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS account_type;
