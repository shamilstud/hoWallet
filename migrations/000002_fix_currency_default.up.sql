-- Fix existing accounts with non-standard currencies to USD
UPDATE accounts SET currency = 'USD' WHERE currency NOT IN ('USD', 'EUR', 'GBP', 'RUB', 'KZT', 'CNY', 'JPY', 'TRY', 'UAH');

-- Update the default for new accounts
ALTER TABLE accounts ALTER COLUMN currency SET DEFAULT 'USD';
