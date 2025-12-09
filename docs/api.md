# Crypto Payment Gateway API Documentation

Base URL: `http://localhost:8080`

## 1. Authentication
- **Header**: `X-API-Key: <YOUR_API_KEY>`

---

## 2. Endpoints

### 2.1 Create Merchant (Managed)
Registers a new merchant and auto-generates a secure wallet (Native SegWit compliant).

- **URL**: `/merchants`
- **Method**: `POST`
- **Auth**: None

#### Request Body
```json
{
  "name": "My Tech Store",
  "payout_address_btc": "bc1q...",
  "payout_address_ltc": "ltc1...",
  "payout_address_eth": "0x...",
  "webhook_url": "https://example.com/webhook"
}
```

#### Response (200 OK)
```json
{
  "id": "550e8400-...",
  "name": "My Tech Store",
  "api_key": "3f889988... (Keep this safe!)",
  "mnemonic": "abandon ability able about above absent absorb abstract absurd abuse access accident",
  "warning": "WRITE THIS DOWN. WE DO NOT DISPLAY IT AGAIN. YOU OWN THIS WALLET.",
  "created_at": "2023-10-27T10:00:00Z"
}
```

---

### 2.2 Generate Wallets (Unified)
Generates unique deposit addresses (BTC & LTC) for a user by deriving them from your xpubs.

- **URL**: `/wallets`
- **Method**: `POST`
- **Auth**: Required (`X-API-Key`)

#### Request Body
```json
{
  "user_id": "user_456"   // Your internal user ID
}
```

#### Response (200 OK)
```json
{
  "user_id": "user_456",
  "merchant_id": "550e8400-...",
  "wallets": [
    {
      "address": "bc1q...",
      "coin": "BTC",
      "derivation_index": 1,
      "status": "active"
    },
    {
      "address": "ltc1...",
      "coin": "LTC",
      "derivation_index": 1,
      "status": "active"
    }
  ]
}
```

---

## 3. How it Works (HD Wallets)

1.  **No Fees**: Funds sent to these addresses appear immediately in your wallet.
2.  **No Sweeping**: We do not move the funds. We just track them.
3.  **Privacy**: We do not have your private keys.

---

## 4. Errors
- `400 Bad Request`: Invalid JSON or missing xpub.
- `401 Unauthorized`: Invalid API Key.
