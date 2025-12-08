# Project Context: Crypto Payment Gateway (HD Wallet Edition)

## 1. Project Overview
**Goal**: Professional, Non-Custodial crypto payment gateway.
**Architecture**: HD Wallet (BIP32) based. The server tracks funds but does not hold keys.

## 2. Core Features
1.  **Multi-Coin**: BTC, LTC, ETH.
2.  **Non-Custodial**: Merchant provides `xpub`, server derives `xpub/0/i` (BTC/LTC) or `xpub` derived addresses (ETH).
3.  **Zero Fees**: Funds go directly to merchant.
4.  **Monitoring**: Monitors blockchain for deposits to derived addresses.

## 3. Directory Structure
```text
/root/gateway/
├── cmd/
│   └── payment-gateway/
├── docs/                # API & Manuals
├── internal/
│   ├── api/             # Handlers (Merchant, Wallets)
│   ├── config/          # Env (DB, RPC, Testnet)
│   ├── crypto/          # Core Logic
│   │   ├── btc/         # Wallet (Derivation), Client
│   │   ├── ltc/         # Wallet (Derivation), Client (Re-uses btcsuite)
│   │   └── eth/         # Wallet, Client
│   ├── db/              # Postgres (Merchants, Addresses, Tx)
│   └── monitor/         # Monitors (BTC, LTC, ETH)
└── go.mod
```

## 4. Derived Schema
- **merchants**: Stores `xpub` and `current_index`.
- **deposit_addresses**: Stores `derivation_index` mapping to users.
- **transactions**: Logs `detected` deposits.

## 5. Deployment
- Requires: PostgreSQL, Bitcoin Node, Litecoin Node.
- Build: `go build ./...`
