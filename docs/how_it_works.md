# How It Works: The "Sweeping" Mechanics

This guide explains the "Magic" of the payment gateway in simple terms.

## 1. What happens when a user deposits?

1.  **Detection**: Your `monitor` (background worker) is constantly asking the Blockchain: *"Hey, are there any new transactions for my addresses?"*
2.  **Match**: When user `001` sends **0.01 BTC**, the monitor sees it and saves it to the database as `detected`.
3.  **Confirmation**: The monitor waits for block confirmations (e.g., 1 block for BTC) to ensure the money is real. Status changes to `confirmed`.

## 2. The Sweeper (The "Forwarding" Process)

Once confirmed, the **Sweeper** wakes up to move the money to YOU (The Merchant).

### "Do we pay fees?" -> YES.
Blockchain transactions are never free. 
Since **our server** is creating the transaction to move funds from the `User's Deposit Address` -> `Merchant's Base Address`, **we must pay the miner fee**.

### How it works mathematically:
Let's say the fee is currently `0.0001 BTC` (approx $3).

| Item | Amount |
| :--- | :--- |
| **User Deposited** | `0.0100 BTC` |
| **Network Fee (Gas)** | `-0.0001 BTC` |
| **Merchant Receives**| `0.0099 BTC` |

**Who pays?** The user effectively pays it, because we deduct it from the total pile before sending it to you. You receive the "Net" amount.

### Why is this necessary?
The Deposit Address is a real wallet on the blockchain. To move funds *out* of it, a transaction must be broadcast, and miners demand a fee to process it.

## 3. Deployment & Production Logic

In a professional setup, the Sweeper logic (`sweeper.go`) does this:
1.  **Calculate Fee**: Ask the network for the current "Fee Per Byte" to ensure fast delivery.
2.  **Sign**: Decrypt the private key (in memory for 10ms), sign the transaction.
3.  **Broadcast**: Push it to the network.
4.  **Update DB**: Mark the detected transaction as `swept`.

---

## 4. Quick Start Guide (Professional Mode)

### Step 1: Run it
```bash
./payment-gateway
```

### Step 2: Create Your Merchant Account
One time setup to get your API Key.
```bash
curl -X POST http://localhost:8080/merchants \
  -d '{"name": "My Business", "btc_base_address": "bc1q...", "ltc_base_address": "ltc1..."}'
```

### Step 3: Integrate with your Website
When User `001` clicks "Deposit Crypto":
**Request:**
```bash
curl -X POST http://localhost:8080/wallets \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"user_id": "001"}'
```

**Response (You show this to user):**
```json
{
  "user_id": "001",
  "wallets": [
    { "coin": "BTC", "address": "bc1qUserAddress..." },
    { "coin": "LTC", "address": "ltc1UserAddress..." }
  ]
}
```
*That's it! Show these QR codes to your user. The system handles the rest.*
