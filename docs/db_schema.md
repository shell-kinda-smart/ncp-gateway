# 🗄️ Database Schema Documentation

The system uses a **PostgreSQL** relational database designed for high integrity, scalability, and auditability.

## 📊 Entity Relationship Diagram (ERD)

```text
+----------------+          +-------------------+
|   MERCHANTS    |          | DEPOSIT_ADDRESSES |
+----------------+          +-------------------+
| id (PK)        |--1:N---->| id (PK)           |
| name           |          | address (UK)      |
| api_key (UK)   |          | coin              |
| btc_xpub       |          | derivation_index  |
| ltc_xpub       |          | user_id           |
| eth_xpub       |          | address_type      |
| settings       |          | status            |
| current_index  |          +-------------------+
+----------------+

       | 1:N
       v

+----------------+
|  TRANSACTIONS  |
+----------------+
| id (PK)        |
| tx_id          |
| coin           |
| amount         |
| to_address     |
| status         |
| merchant_id(FK)|
+----------------+

+----------------+
|    APP_STATE   |
+----------------+
| key (PK)       |
| value          |
+----------------+
```

## 📝 Table Details

### 1. `merchants`
The core identity. Stores the **Extended Public Keys (xpubs)**.
*   **Security**: No private keys are stored here. Only specific `xpubs`.
*   **Scalability**: `current_index` tracks how many addresses have been generated to ensure we never reuse an address.
*   **Flexibility**: `settings` is a JSONB column, allowing us to add features (like multiple webhooks, slack alerts, email configs) without changing the schema.

### 2. `deposit_addresses`
Links a specific blockchain address to a specific Merchant and Customer.
*   **Unique Address Per User**: Every time you call `/wallets`, we increment the merchant's `current_index` and generate a fresh address at `m/0/new_index`.
*   **Traceability**: `user_id` allows you to link a deposit back to your internal Order ID or User ID.

### 3. `transactions`
An immutable log of all detected blockchain events.
*   **Audit Trail**: Every deposit detected by the Monitor is logged here before the Webhook is fired.
*   ** Deduplication**: The combination of `tx_id` and `to_address` ensures we process a deposit only once.

### 4. `app_state`
The brain of the background workers.
*   **Bulletproof Monitoring**: Stores the last block height processed for each coin. If the server crashes and restarts 5 hours later, it reads this value and "catches up" on all missing blocks automatically. 
