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
*   **SegWit Support**: Generates efficient Native SegWit (`bc1q...`, `ltc1...`) addresses.
*   **Privacy Focused**: Server only knows the `xpub`; it cannot spend funds.
*   **Scalable**: Built on Go and PostgreSQL.

## 🏗 Architecture

The system uses **Hierarchical Deterministic (HD) Keys**.
1.  **Merchant** provides an Extended Public Key (`xpub`).
2.  **Gateway** derives a unique child address (`m/0/i`) for each customer.
3.  **Customer** pays the address.
4.  **Gateway** detects the transaction via Blockchain RPC and notifies the merchant.
5.  **Merchant** sees funds immediately in their own wallet.

---

## 🛠 Prerequisites

*   **Go**: 1.23 or higher.
*   **PostgreSQL**: Database for tracking addresses and transactions.
*   **Bitcoin Core**: Node with RPC enabled (`server=1`, `txindex=1` recommended).
*   **Litecoin Core**: Node with RPC enabled.

## 📦 Installation

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

---

## 🔌 Usage API

### 1. Register Merchant
**POST** `/merchants`
```json
{
  "name": "My Store",
  "btc_xpub": "xpub...",
  "ltc_xpub": "xpub..."
}
```

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
