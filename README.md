# Crypto Payment Gateway (HD Wallet Edition)

[![Go Version](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-active-success.svg)]()

A professional, self-hosted, non-custodial cryptocurrency payment gateway supporting **Bitcoin (BTC)** and **Litecoin (LTC)**.  
Built for privacy, security, and zero fees using **HD Wallets (BIP32)**.

---

## 📖 Table of Contents
- [Features](#-features)
- [Architecture](#-architecture)
- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage (API)](#-usage-api)
- [GitHub Setup](#-github-setup)

---

## 🚀 Features

*   **Non-Custodial**: You hold the keys. Funds settle directly to your hardware/software wallet.
*   **Zero Fees**: No sweeping transactions. No miner fees paid by the gateway.
*   **SegWit & Legacy**: Smart auto-detection (`bc1q...` vs `1...`) based on your wallet preference.
*   **Privacy Focused**: Server only knows the `xpub`; it cannot spend funds.
*   **Real-Time Prices**: Webhooks include USD conversion rates (via CoinGecko).
*   **Professional Webhooks**: Clean, rich JSON payloads with complete metadata.
*   **Scalable**: Built on Go and PostgreSQL.

## 🏗 Architecture (Pure Managed)

The gateway operates in **Pure Managed Mode**:

*   **Gateway** generates a 12-word Seed (BIP39).
*   **Gateway** holds encryption keys (AES-GCM).
*   **Gateway** tracks your balance (Watch-Only) on your node.
*   **You** trigger Manual Withdrawals to move funds to Cold Storage.

---

## 🛠 Prerequisites

*   **Go**: 1.23 or higher.
*   **PostgreSQL**: Database for tracking addresses and transactions.
*   **Bitcoin Core**: Node with RPC enabled (`server=1`, `txindex=1` recommended).
*   **Litecoin Core**: Node with RPC enabled.
*   **Ethereum**: Geth or Infura/Alchemy RPC URL.

## �️ Troubleshooting & Manual Run

If you see `go: command not found`, you need to add Go to your path:
```bash
export PATH=$PATH:/root/gateway/.go_dist/bin
```

Then you can build and run manually:
```bash
go build -o payment-gateway ./cmd/payment-gateway
./payment-gateway
```

## �📦 Installation

1.  **Clone the repository** (or use the provided source):
    ```bash
    git clone https://github.com/YOUR_USERNAME/crypto-gateway.git
    cd crypto-gateway
    ```

2.  **Build the binary**:
    ```bash
    go build -o payment-gateway ./cmd/payment-gateway
    ```

3.  **Run the server**:
    ```bash
    ./payment-gateway
    ```
    *The application will automatically apply database migrations on startup.*

---

## ⚙ Configuration

Create a `.env` file in the root directory:

```properties
# Server
PORT=8080
TESTNET=true

# Database
DATABASE_URL=postgres://user:password@localhost:5432/gateway_db?sslmode=disable

# Bitcoin Node
BTC_RPC_URL=http://localhost:8332
BTC_RPC_USER=rpcuser
BTC_RPC_PASS=rpcpassword

# Litecoin Node
LTC_RPC_URL=http://localhost:9332
LTC_RPC_USER=rpcuser
LTC_RPC_PASS=rpcpassword
```

### 💡 Best Free RPC Providers
Don't want to run your own node? Use the **Free Tier** of these providers:
1.  **[GetBlock.io](https://getblock.io/)** (40k req/day - Supports BTC & LTC).
2.  **[NOWNodes.io](https://nownodes.io/)** (100k req/day - Supports BTC & LTC).

---

## 🔌 Usage API

### 1. Register Merchant (Managed)
**POST** `/merchants`
```json
{
  "name": "My Store",
  "payout_address_btc": "bc1q...",
  "payout_address_ltc": "ltc1...",
  "payout_address_eth": "0x..."
}
```
**Response** will include your **API Key** and **Secret Mnemonic**. Save them!

### 2. Generate Payment Wallets
**POST** `/wallets`  
**Header**: `X-API-Key: <YOUR_API_KEY>`
```json
{
  "user_id": "customer_01"
}
```

---

## 🐙 GitHub Setup

If you have just initialized this project locally and want to push it to a new private GitHub repository:

1.  **Create a new Repository** on GitHub (do not initialize with README/License).
2.  **Run these commands** in the project root:

```bash
# 1. Initialize Git (if not done)
git init

# 2. Add your GitHub Remote
git remote add origin https://github.com/YOUR_USERNAME/crypto-gateway.git

# 3. Rename branch to main
git branch -M main

# 4. Push
git push -u origin main
```
