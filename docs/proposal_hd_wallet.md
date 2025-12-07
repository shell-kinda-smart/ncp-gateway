# Architecture Proposal: HD Wallets (xpub)

The user asked: *"Can we use subaddresses so we don't need to transfer (sweep) and pay gas?"*

**The Answer: YES.**

To do this, we must switch from a **"Deposit & Sweep"** model to an **"HD Wallet (xpub)"** model.

## key Differences

| Feature | Current (sweeping) | Proposed (HD Wallet) |
| :--- | :--- | :--- |
| **Merchant Setup** | Provide Single Address (`bc1q...`) | Provide Extended Public Key (`xpub...` or `zpub...`) |
| **User Address** | Randomly generated Key Pair | Derived from Merchant's xpub (`m/84'/0'/0'/0/1`) |
| **Private Keys** | **We hold them** (Encrypted DB) | **We NEVER see them** (Merchant holds Master Seed) |
| **Funds Location** | Sits in DB-wallet until swept | **Instantly in Merchant's Wallet** |
| **Gas Fees** | **YES** (1 Tx to sweep) | **NO** (Zero fees to receive) |
| **Consolidation** | Automatic (We send to Base) | Manual (Merchant manages UTXOs in their wallet app) |

## How it works for the Merchant

1.  Merchant opens their wallet (e.g., Electrum, Trezor, Ledger).
2.  Copies their **Master Public Key (zpub/xpub)**.
3.  Pastes it into our Gateway.
4.  **User 001** requests a deposit address.
5.  Gateway derives `Address #1` from the xpub and shows it to User 001.
6.  User 001 pays 1 BTC.
7.  **Magic**: The Merchant sees "1 BTC" appear in their Electrum/Ledger balance **immediately**.
8.  **No Sweeping Required**. The funds are already in a key derived from their seed.

## Required Changes

1.  **Database**:
    - Merchants: Replace `btc_base_address` with `btc_xpub`.
    - DepositAddresses: Remove `encrypted_private_key`. Add `derivation_path` (e.g., `0/1`).
2.  **Logic**:
    - **Remove**: `encryption.go` (We don't need it anymore! Huge security win).
    - **Remove**: `sweeper.go` (We don't sweep anymore).
    - **Update**: `wallet.go` to use `btcutil/hdkeychain` to derive addresses from xpub.
    - **Update**: `monitor.go` remains mostly the same (just watching addresses).

## Recommendation

This is the **Professional Standard** for crypto payment gateways (like BTCPay Server).
- **Pros**: Zero Fees for us/merchant on receive, Non-Custodial (Hackers can't steal funds from our DB), Merchant is in full control.
- **Cons**: Merchant must verify their xpub correctly (slightly more complex setup for them).

**Shall we proceed with this refactor?**
