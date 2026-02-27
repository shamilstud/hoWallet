# hoWallet

Personal & shared finance tracker — your alternative to Buxfer.

## Features

- **Accounts**: Track bank cards, deposits, cash in one place
- **Transactions**: Income, expenses, transfers between accounts
- **Shared Wallets**: Invite your partner — both can view & manage finances together
- **CSV Export**: Buxfer-compatible CSV export
- **Auth**: JWT-based authentication with refresh tokens

## Tech Stack

- **Backend**: Go 1.21 + chi + sqlc + pgx
- **Database**: PostgreSQL 16
- **Frontend**: Next.js 14 + TypeScript + Tailwind CSS
- **Containerization**: Docker Compose

## Quick Start

```bash
# 1. Clone and configure
cp .env.example .env
# Edit .env with your settings

# 2. Start everything
docker compose up -d --build

# 3. API is at http://localhost:8080
# 4. Frontend is at http://localhost:3000
```

## Development

```bash
# Install tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Generate sqlc code after changing SQL queries
make sqlc

# Run migrations manually
make migrate-up

# Run tests
make test
```

## API Endpoints

### Auth
- `POST /auth/register` — Register a new user
- `POST /auth/login` — Login
- `POST /auth/refresh` — Refresh access token
- `POST /auth/logout` — Logout (requires auth)

### Households
- `POST /api/households` — Create a wallet group
- `GET /api/households` — List your wallet groups
- `GET /api/households/:id/members` — List members
- `POST /api/households/:id/invite` — Invite by email
- `DELETE /api/households/:id/members/:userId` — Remove member
- `POST /api/invitations/:token/accept` — Accept invitation

### Accounts (requires `X-Household-ID` header)
- `POST /api/accounts` — Create account
- `GET /api/accounts` — List accounts
- `GET /api/accounts/:id` — Get account
- `PUT /api/accounts/:id` — Update account
- `DELETE /api/accounts/:id` — Delete account

### Transactions (requires `X-Household-ID` header)
- `POST /api/transactions` — Create transaction
- `GET /api/transactions` — List (filters: `from`, `to`, `type`, `account_id`, `limit`, `offset`)
- `GET /api/transactions/:id` — Get transaction
- `PUT /api/transactions/:id` — Update transaction
- `DELETE /api/transactions/:id` — Delete transaction

### Export (requires `X-Household-ID` header)
- `GET /api/export/csv` — Export as Buxfer-compatible CSV (filters: `from`, `to`)
