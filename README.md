# NCP Gateway

### Your keys. Your settlement. No middleman custody.

[![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Required-4169E1?logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Bitcoin](https://img.shields.io/badge/Bitcoin-BTC-F7931A?logo=bitcoin&logoColor=white)](https://bitcoin.org/)
[![Litecoin](https://img.shields.io/badge/Litecoin-LTC-345D9D?logo=litecoin&logoColor=white)](https://litecoin.org/)
[![Status](https://img.shields.io/badge/Status-Prototype-8A2BE2)](#project-status)

**NCP Gateway** is a self-hosted, non-custodial cryptocurrency payment backend for Bitcoin and Litecoin. It derives a fresh Native SegWit deposit address for each customer from a merchant's HD wallet extended public key, then monitors blockchain nodes for incoming payments.

Funds travel directly from the customer to addresses controlled by the merchant's wallet. The gateway never receives a private key, cannot sign transactions, and does not sweep funds.

> [!IMPORTANT]
> This repository is an early prototype and has not been security audited. Its current startup migration drops and recreates application tables. Use testnet and disposable development data only.

## Why NCP Gateway?

- **Non-custodial by design** — the service stores public wallet data, never spending keys.
- **Direct settlement** — customer payments arrive in the merchant's own HD wallet.
- **Unique payment addresses** — deterministic `m/0/i` child addresses improve reconciliation and reduce address reuse.
- **BTC + LTC support** — derives Native SegWit Bitcoin and Litecoin addresses.
- **API-first integration** — register merchants and create customer deposit addresses over HTTP.
- **Self-hosted infrastructure** — run the Go service, PostgreSQL, and your own blockchain RPC nodes.
- **Testnet mode** — develop without risking real funds.

## How it works

```mermaid
flowchart LR
    M[Merchant HD wallet] -->|BTC/LTC extended public keys| G[NCP Gateway]
    G -->|Derive m/0/i| A[Unique SegWit address]
    A --> C[Customer payment]
    C --> W[Merchant-controlled wallet]
    N[Bitcoin & Litecoin nodes] -->|Block data via RPC| G
    G -->|Addresses and detected deposits| P[(PostgreSQL)]
```

1. A merchant registers one or both supported extended public keys.
2. NCP Gateway returns a merchant API key.
3. The merchant requests deposit addresses for an internal customer or order ID.
4. The gateway atomically advances the derivation index and derives `m/0/i` addresses.
5. The customer pays the generated address directly.
6. Background BTC and LTC monitors inspect blockchain data and record matching deposits.

## Architecture

```text
cmd/payment-gateway/       Application entry point and HTTP server
internal/api/              Merchant and wallet HTTP handlers
internal/config/           Environment-based configuration
internal/crypto/btc/       Bitcoin RPC, HD derivation and monitoring
internal/crypto/ltc/       Litecoin RPC, HD derivation and monitoring
internal/db/               PostgreSQL schema and data access
docs/                      API and architecture notes
```

The service uses Go's standard `net/http` package, PostgreSQL via `lib/pq`, and the `btcsuite`/`ltcsuite` libraries for chain-specific key and transaction handling.

## Quick start

### Requirements

- Go 1.23 or newer
- PostgreSQL
- Bitcoin Core with RPC enabled
- Litecoin Core with RPC enabled
- Account-level extended public keys for the network selected by `TESTNET`

### 1. Clone and configure

```bash
git clone https://github.com/shell-kinda-smart/ncp-gateway.git
cd ncp-gateway
```

Create `.env` in the project root:

```dotenv
PORT=8080
TESTNET=true

DATABASE_URL=postgres://gateway:password@localhost:5432/ncp_gateway?sslmode=disable

BTC_RPC_URL=http://127.0.0.1:18332
BTC_RPC_USER=bitcoin_rpc_user
BTC_RPC_PASS=bitcoin_rpc_password

LTC_RPC_URL=http://127.0.0.1:19332
LTC_RPC_USER=litecoin_rpc_user
LTC_RPC_PASS=litecoin_rpc_password
```

Keep `.env`, API keys, RPC credentials, and private wallet material out of version control.

### 2. Build and run

```bash
go build -o ncp-gateway ./cmd/payment-gateway
./ncp-gateway
```

For development:

```bash
go run ./cmd/payment-gateway
```

The HTTP server listens on `http://localhost:8080` unless `PORT` is changed.

## API overview

### Register a merchant

```http
POST /merchants
Content-Type: application/json
```

```json
{
  "name": "Satoshi Supply Co.",
  "btc_xpub": "xpub...",
  "ltc_xpub": "xpub..."
}
```

The response includes the merchant ID and a newly generated API key. Store the API key securely; it authorizes wallet-generation requests.

### Create customer deposit addresses

```http
POST /wallets
Content-Type: application/json
X-API-Key: YOUR_MERCHANT_API_KEY
```

```json
{
  "user_id": "order_2026_001"
}
```

The response contains the derived BTC and/or LTC addresses configured for that merchant, along with their derivation indexes.

See [docs/api.md](docs/api.md) for response shapes and error codes.

## Security model

NCP Gateway is intentionally unable to spend merchant funds:

- It accepts extended **public** keys, not seed phrases or private keys.
- It derives public receiving addresses only.
- It does not construct, sign, sweep, or broadcast outgoing transactions.
- Payment funds settle to addresses recoverable by the merchant's HD wallet.

The service can still observe merchant addresses and payment activity. Protect the database, API keys, RPC endpoints, and network perimeter. Production deployments also require HTTPS, input hardening, rate limiting, secret rotation, observability, backups, and an independent security review.

## Project status

Implemented today:

- Merchant registration with BTC and/or LTC extended public keys
- Cryptographically random merchant API keys
- Atomic derivation-index allocation in PostgreSQL
- Bitcoin and Litecoin Native SegWit address derivation
- API-key-protected address generation
- Background node polling and deposit recording
- Mainnet/testnet network selection

Before production use:

- Replace destructive startup schema resets with versioned migrations
- Persist processed block heights and handle chain reorganizations
- Enforce idempotent transaction recording and confirmation updates
- Add signed webhooks with retries and delivery history
- Hash stored API keys and add rotation/revocation
- Add request validation, rate limiting, metrics, tests, and audit logging
- Complete a professional security audit

## Documentation

- [API reference](docs/api.md)
- [HD wallet architecture proposal](docs/proposal_hd_wallet.md)
- [Project context](context.md)

## Disclaimer

Cryptocurrency payment infrastructure is security-critical software. Test thoroughly on testnet and obtain professional review before processing real funds. You are responsible for node security, wallet backups, regulatory compliance, and operational risk.
