# Database Setup Guide (PostgreSQL)

This guide explains how to set up the PostgreSQL database for the Crypto Payment Gateway manually.

## 1. Install PostgreSQL
**Ubuntu/Debian**:
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

## 2. Access the Database Shell
```bash
sudo -u postgres psql
```

## 3. Create Database & User
Run these SQL commands inside the `psql` shell:

```sql
-- 1. Create User
CREATE USER gateway_user WITH PASSWORD 'secure_password';

-- 2. Create Database
CREATE DATABASE gateway_db OWNER gateway_user;

-- 3. Grant Privileges
GRANT ALL PRIVILEGES ON DATABASE gateway_db TO gateway_user;
```

## 4. Connect & Create Tables
You can let the application auto-create tables (it runs migrations on startup), OR you can create them manually for control.

**Manual SQL Schema**:
```sql
\c gateway_db

-- Merchants Table
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    api_key TEXT UNIQUE NOT NULL,
    btc_xpub TEXT,
    ltc_xpub TEXT,
    eth_xpub TEXT,
    settings JSONB DEFAULT '{}',
    current_index INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Deposit Addresses Table
CREATE TABLE deposit_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address TEXT UNIQUE NOT NULL,
    coin TEXT NOT NULL,
    merchant_id UUID REFERENCES merchants(id),
    derivation_index INT NOT NULL,
    user_id TEXT,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Transactions Table
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tx_id TEXT NOT NULL,
    coin TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    from_address TEXT,
    to_address TEXT,
    confirmations INT DEFAULT 0,
    status TEXT DEFAULT 'detected',
    merchant_id UUID REFERENCES merchants(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- App State (For Bulletproof Monitoring)
CREATE TABLE app_state (
    key TEXT PRIMARY KEY,
    value TEXT
);

-- Indexes for Performance
CREATE INDEX idx_deposit_address ON deposit_addresses(address);
CREATE INDEX idx_tx_to_address ON transactions(to_address);
```

## 5. Update .env
Update your `.env` file to match the user/password you created:
```bash
DATABASE_URL=postgres://gateway_user:secure_password@localhost:5432/gateway_db?sslmode=disable
```
