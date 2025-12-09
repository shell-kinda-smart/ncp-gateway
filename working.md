# ⚙️ How It Works: System Internals

This document explains the "Magic" behind the Payment Gateway: how we secure keys, talk to the blockchain, and process money.

## 1. 🔐 The Vault (Security & Encryption)

We use a **Custodial** model with **High-Security Encryption**.

### A. Generation
1.  **Entropy**: We generate a random 12-word BIP39 Mnemonic (Seed).
2.  **Derivation**: From this one seed, we mathematically derive keys for all coins:
    *   **BTC**: `m/84'/0'/0'` (Native SegWit `bc1q...`)
    *   **LTC**: `m/44'/2'/0'` (Standard `ltc1...`)
    *   **ETH**: `m/44'/60'/0'` (Standard `0x...`)

### B. Encryption (AES-GCM)
We **NEVER** store the Mnemonic in plain text.
1.  **Master Key**: The system loads a 32-byte `MASTER_KEY` from the `.env` file (computer memory only).
2.  **Encrypt**: We use `AES-256-GCM` (Galois/Counter Mode) to encrypt the Mnemonic.
    *   This adds an "Authentication Tag" so we verify the data hasn't been tampered with.
3.  **Storage**: The resulting **Encrypted Hex String** is saved in the **PostgreSQL** database (`merchants` table).

**Result**: If hackers steal the database, they get garbage text. They need the `MASTER_KEY` (from the server environment) to read it.

---

## 2. 📡 Blockchain Connection (RPC)

We do not rely on 3rd party APIs (like Etherscan). We talk directly to the **Blockchain Node**.

*   **URL**: Defined in `.env` (e.g., `BTC_RPC_URL=http://localhost:8332`).
*   **Protocol**: JSON-RPC over HTTP.

### How we "Watch" (Watch-Only Wallet)
We don't keep the private key on the Bitcoin/Litecoin Node (for security).
1.  **Import**: When you create a merchant, we send an `importdescriptors` command to the Node.
2.  **What we send**: The **Public Key** (xpub).
3.  **Effect**: The Node now scans the blockchain for any money sent to that xpub, but it **cannot spend it**. It just watches.

---

## 3. 🔄 The Cycle: From Deposit to Withdrawal

### Step 1: The Deposit
1.  **User** sends BTC to your address.
2.  **Miner** includes it in a Block.
3.  **Gateway Monitor** (running in background) polls the Node:
    *   `getblockchaininfo` (Check for new height).
    *   `getblock` (Download full block data).
4.  **Match**: The Gateway checks every transaction in the block.
    *   *"Does this output address match one of our Merchant addresses?"*
5.  **Record**: If match found -> Save to `transactions` table in DB.
6.  **Webhook**: Fire a `POST` request to your `webhook_url`.

### Step 2: The Withdrawal (Manual)
You want to move funds to Cold Storage.
1.  **Request**: You send `POST /withdraw` with your API Key.
2.  **Unlock**:
    *   Gateway fetches Encrypted Mnemonic from DB.
    *   Decrypts it using `MASTER_KEY`.
3.  **Construct**:
    *   Gateway asks Node: "List Unspent Outputs (UTXOs) for this watch-only wallet".
    *   Gateway builds the transaction locally (Client-side signing).
4.  **Sign**: Gateway signs the transaction using the decrypted Mnemonic.
5.  **Broadcast**: Gateway sends the **Signed Hex** to the Node (`sendrawtransaction`).
6.  **Done**: Funds move to your destination.

---

## 4. 📂 Directory of Important Files

*   **Encryption Logic**: `internal/security/encryption.go`
*   **Key Generation**: `internal/crypto/generator/generator.go`
*   **Bitcoin Client**: `internal/crypto/btc/client.go`
*   **Withdrawal Handler**: `internal/api/handlers.go` (Func `WithdrawHandler`)
