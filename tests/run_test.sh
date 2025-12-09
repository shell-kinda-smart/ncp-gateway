#!/bin/bash
set -e

echo "🚀 Starting Full System Verification..."

# Ensure Go is in PATH (Fix for user environment)
export PATH=$PATH:/root/gateway/.go_dist/bin:/usr/local/go/bin

# 1. Setup Database & Test Data
echo "📦 Setting up Test Data (Merchant, Addresses, Webhooks)..."
cd .. # Go to root
go run tests/setup/setup_merchant.go

# 2. Build Gateway
echo "🔨 Building Gateway..."
go build -o payment-gateway ./cmd/payment-gateway

# 3. Run Gateway
echo "🟢 Running Gateway (Press Ctrl+C to stop)..."
./payment-gateway
