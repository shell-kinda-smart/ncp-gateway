# Crypto Payment Gateway (Self-Hosted)

A professional, non-custodial Bitcoin (BTC) and Litecoin (LTC) payment gateway written in Go.

## 🚀 Features

*   **Non-Custodial**: We never hold your private keys. Funds go directly to your hardware/software wallet.
*   **Zero Fees**: No sweeping fees. Partial deposits are credited instantly to your wallet.
*   **Multi-Coin**: Native SegWit support for Bitcoin and Litecoin.
*   **HD Wallets**: Uses BIP32/BIP84 (xpub) to derive unique addresses for every customer.
*   **Database**: robust PostgreSQL backend.

## 🛠 Prerequisites

1.  **Go 1.23+**: [Install Go](https://go.dev/dl/)
2.  **PostgreSQL**: A running database server.
3.  **Bitcoin Core**: A synced node with RPC enabled (`server=1`).
4.  **Litecoin Core**: A synced node with RPC enabled (`server=1`).

## ⚙️ Configuration

Create a `.env` file in the project root:

```bash
# Database
DATABASE_URL=postgres://user:password@localhost:5432/gateway_db?sslmode=disable

# Server Config
PORT=8080
TESTNET=true # Set to 'false' for Mainnet

# Bitcoin Node RPC
BTC_RPC_URL=http://localhost:8332
BTC_RPC_USER=rpcuser
BTC_RPC_PASS=rpcpassword

# Litecoin Node RPC
LTC_RPC_URL=http://localhost:9332
LTC_RPC_USER=rpcuser
LTC_RPC_PASS=rpcpassword
```

## 🏗 Build & Run

### 1. Build
```bash
go build -o payment-gateway ./cmd/payment-gateway
```

### 2. Run
```bash
./payment-gateway
```
*The server will start on port 8080 and automatically create necessary database tables.*

## 🔌 API Usage

### 1. Register Merchant
One-time setup. Provide your **Extended Public Keys (xpubs)** from your Ledger/Trezor/Electrum.

```bash
curl -X POST http://localhost:8080/merchants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Store",
    "btc_xpub": "xpub6C...",
    "ltc_xpub": "xpub6B..."
  }'
```
**Response:** Returns your `api_key`. **Save this!**

### 2. Generate Payment Addresses
When a user wants to deposit, call this endpoint.

```bash
curl -X POST http://localhost:8080/wallets \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{"user_id": "user_123"}'
```

**Response:**
```json
{
  "wallets": [
    { "coin": "BTC", "address": "bc1q..." },
    { "coin": "LTC", "address": "ltc1..." }
  ]
}
```

## 📚 Documentation
Check the `docs/` folder for more details:
*   [API Documentation](docs/api.md)
*   [How It Works (Refactored)](docs/proposal_hd_wallet.md)
