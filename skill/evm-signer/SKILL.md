---
name: evm-signer
description: Secure USDC and EVM transaction signing for AI agents. Use when you need to send USDC payments, native ETH transfers, or ERC20 tokens through a policy-controlled local signer with risk control rules.
---

# EVM Signer Skill

> **Built for the OpenClaw USDC Hackathon** — Enabling AI agents to handle USDC payments securely, with policy-based guardrails.

## The Problem

AI agents are increasingly executing on-chain transactions — **USDC payments**, swaps, NFT mints, DeFi interactions. But there's a fundamental security problem:

**You can't give an AI agent your private key.**

If you do:
- The agent could sign anything (no guardrails)
- A prompt injection could drain your wallet
- There's no audit trail or policy enforcement
- One bug = total loss of funds

Current solutions either give agents full key access (dangerous) or require human approval for every transaction (defeats the purpose of autonomy).

## The Solution

**EVM Signer** is a local signing service that sits between your AI agent and your private keys:

```
AI Agent                    EVM Signer                 Blockchain
   │                            │                          │
   │  "Sign this tx"            │                          │
   ├───────────────────────────►│                          │
   │                            │  1. Check IP whitelist   │
   │                            │  2. Match against rules  │
   │                            │  3. Sign if allowed      │
   │   Signed raw tx            │                          │
   │◄───────────────────────────┤                          │
   │                            │                          │
   │  Broadcast via RPC ────────┼─────────────────────────►│
```

**Key insight**: The agent never sees the private key. It only receives signed transactions — and only for operations that match your pre-defined rules.

## Three Security Layers

### 1. Key Isolation
Private keys are stored in encrypted keystores (go-ethereum standard). They're decrypted once at service startup and exist only in memory. The HTTP API never exposes keys — only signed outputs.

### 2. Policy Engine
Every transaction must match at least one rule in `rule.json`. No match = no signature. Rules can constrain:
- Recipient addresses (whitelist)
- Transfer amounts (caps)
- Token contracts (only specific ERC20s)
- Function selectors (only specific contract calls)

### 3. Network Isolation
The service binds to localhost by default. IP whitelist ensures only your agent can access it. For production: run behind VPN or in a trusted enclave.

---

## What This Skill Enables

### Task 1: Send USDC Payments

**USDC on Ethereum Mainnet:**
```bash
python3 send_transaction.py \
  --chain-id 1 \
  --from 0xYourWallet \
  --token 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 \
  --to 0xRecipient \
  --amount 10000000
```
Sends 10 USDC on Ethereum (USDC has 6 decimals, so 10000000 = 10 USDC).

**USDC on Base:**
```bash
python3 send_transaction.py \
  --chain-id 8453 \
  --from 0xYourWallet \
  --token 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913 \
  --to 0xRecipient \
  --amount 50000000
```
Sends 50 USDC on Base — Circle's native USDC deployment.

**USDC on Polygon:**
```bash
python3 send_transaction.py \
  --chain-id 137 \
  --from 0xYourWallet \
  --token 0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359 \
  --to 0xRecipient \
  --amount 100000000
```
Sends 100 USDC on Polygon.

**USDC on Arbitrum:**
```bash
python3 send_transaction.py \
  --chain-id 42161 \
  --from 0xYourWallet \
  --token 0xaf88d065e77c8cC2239327C5EDb3A432268e5831 \
  --to 0xRecipient \
  --amount 25000000
```
Sends 25 USDC on Arbitrum.

### Task 2: Send Native ETH

**ETH Transfer:**
```bash
python3 send_transaction.py \
  --chain-id 1 \
  --from 0xYourWallet \
  --to 0xRecipient \
  --value 1000000000000000000
```
Sends 1 ETH on Ethereum mainnet.

**Dry Run (sign without broadcast):**
```bash
python3 send_transaction.py \
  --chain-id 8453 \
  --from 0xYourWallet \
  --token 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913 \
  --to 0xRecipient \
  --amount 10000000 \
  --dry-run
```
Returns signed USDC transaction without broadcasting. Useful for verification before committing.

### USDC Contract Addresses

| Chain | Chain ID | USDC Address |
|-------|----------|--------------|
| Ethereum | 1 | `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` |
| Base | 8453 | `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913` |
| Polygon | 137 | `0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359` |
| Arbitrum | 42161 | `0xaf88d065e77c8cC2239327C5EDb3A432268e5831` |
| Optimism | 10 | `0x0b2C639c533813f4Aa9D7837CAf62653d097Ff85` |
| Avalanche | 43114 | `0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E` |

**Note:** These are Circle's native USDC deployments, not bridged versions.

### Task 3: Create Risk Control Rules for USDC

Rules live in `conf/rule.json`. Each rule has:
- `name`: Human-readable identifier
- `chain_id`: Which chain this rule applies to
- `conditions`: Array of constraints (ALL must match)

**Example 1: Allow USDC transfers up to $100 on Ethereum**
```json
{
  "name": "usdc_eth_small_transfer",
  "chain_id": 1,
  "conditions": [
    { "field": "to", "symbol": "==", "value": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" },
    { "field": "data_selector", "symbol": "==", "value": "0xa9059cbb" },
    {
      "field": "data_param",
      "symbol": "<=",
      "value": "100000000",
      "abi": "{\"name\":\"transfer\",\"type\":\"function\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}]}",
      "param": "value"
    }
  ]
}
```
Allows USDC transfers up to 100 USDC (100000000 = 100 * 10^6) on Ethereum.

**Example 2: Allow USDC transfers to whitelisted addresses on Base**
```json
{
  "name": "usdc_base_whitelist",
  "chain_id": 8453,
  "conditions": [
    { "field": "to", "symbol": "==", "value": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913" },
    { "field": "data_selector", "symbol": "==", "value": "0xa9059cbb" },
    {
      "field": "data_param",
      "symbol": "in",
      "value": ["0xTrustedAddress1...", "0xTrustedAddress2..."],
      "abi": "{\"name\":\"transfer\",\"type\":\"function\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}]}",
      "param": "to"
    }
  ]
}
```
Only allows USDC transfers to pre-approved recipient addresses on Base.

**Example 3: Allow USDC payments for API services (amount range)**
```json
{
  "name": "usdc_api_payments",
  "chain_id": 42161,
  "conditions": [
    { "field": "to", "symbol": "==", "value": "0xaf88d065e77c8cC2239327C5EDb3A432268e5831" },
    { "field": "data_selector", "symbol": "==", "value": "0xa9059cbb" },
    {
      "field": "data_param",
      "symbol": ">=",
      "value": "100000",
      "abi": "{\"name\":\"transfer\",\"type\":\"function\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}]}",
      "param": "value"
    },
    {
      "field": "data_param",
      "symbol": "<=",
      "value": "10000000",
      "abi": "{\"name\":\"transfer\",\"type\":\"function\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}]}",
      "param": "value"
    }
  ]
}
```
Allows USDC payments between $0.10 and $10 on Arbitrum — perfect for micro-payments to APIs.

**Example 4: Allow ETH gas top-ups (small amounts only)**
```json
{
  "name": "eth_gas_topup",
  "chain_id": 1,
  "conditions": [
    { "field": "value", "symbol": "<=", "value": "10000000000000000" }
  ]
}
```
Allows ETH transfers up to 0.01 ETH for gas top-ups.

---

## Quick Start

### 1. Clone and Build
```bash
git clone https://github.com/sauldataman/evm-signer
cd evm-signer
make build
```

### 2. Configure Keys
Copy example config:
```bash
cp conf/config.yaml.example conf/config.yaml
```

Edit `conf/config.yaml` to add your account:
```yaml
account:
  type: EncryptedMnemonic
  keys:
    0:
      type: EncryptedMnemonic
      key: ".keystore/mnemonic.enc"
      password: "your-password"
      index: "0"
```

### 3. Define Rules
Copy example rules:
```bash
cp conf/rule.json.example conf/rule.json
```

Edit `conf/rule.json` to whitelist allowed operations.

### 4. Start the Signer
```bash
./signer start
```
Service runs at `http://localhost:8080`.

### 5. Test It
```bash
# Check health
curl http://localhost:8080/ping

# Get your address
curl -X POST http://localhost:8080/v1/address \
  -H "Content-Type: application/json" \
  -d '{"chainId": 1, "account": 0}'
```

---

## API Reference

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/ping` | GET | Health check |
| `/v1/address` | POST | Get wallet address by account index |
| `/v1/sign/transaction` | POST | Sign an EVM transaction |
| `/v1/sign/message` | POST | Sign a plain message |
| `/v1/sign/eip712` | POST | Sign EIP-712 typed data |

### Sign Transaction Request
```json
{
  "chainId": 1,
  "account": 0,
  "tx": {
    "to": "0x...",
    "value": "1000000000000000000",
    "gas": "21000",
    "gasPrice": "20000000000",
    "nonce": 5,
    "data": "0x"
  }
}
```

### Response
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "rawTx": "0xf86c...",
    "txHash": "0x..."
  }
}
```

Error codes:
- `1001`: IP not in whitelist
- `1002`: Transaction rejected by rules
- `1003`: Signing failed

---

## Rule Schema Reference

### Condition Fields

| Field | Description | Example |
|-------|-------------|---------|
| `to` | Recipient address | `"0x..."` |
| `value` | Native token amount in wei | `"1000000000000000000"` |
| `data_selector` | First 4 bytes of calldata | `"0xa9059cbb"` (ERC20 transfer) |
| `data_param` | Decoded ABI parameter | Requires `abi` and `param` fields |
| `from` | Sender address | `"0x..."` |

### Comparison Operators

| Symbol | Meaning |
|--------|---------|
| `==` | Equals |
| `<=` | Less than or equal |
| `>=` | Greater than or equal |
| `in` | Value is in array |
| `contains` | String contains |
| `regex` | Regular expression match |

### Common Selectors

| Selector | Function |
|----------|----------|
| `0xa9059cbb` | ERC20 `transfer(address,uint256)` |
| `0x23b872dd` | ERC20 `transferFrom(address,address,uint256)` |
| `0x095ea7b3` | ERC20 `approve(address,uint256)` |
| `0x38ed1739` | Uniswap `swapExactTokensForTokens` |
| `0x7ff36ab5` | Uniswap `swapExactETHForTokens` |

---

## Why This Matters for Agentic Commerce

As AI agents become economic actors — paying for APIs, executing trades, managing treasuries — the question isn't *if* they'll need to sign transactions, but *how to do it safely*.

EVM Signer provides the missing infrastructure:
- **Autonomy with guardrails**: Agents can transact without human approval, within defined limits
- **Defense in depth**: Multiple security layers, no single point of failure
- **Auditability**: Every signing request is logged, every rule is explicit
- **Multi-chain ready**: One signer for all EVM chains

This is how you give an AI agent a wallet without giving it the keys.

---

## Resources

- **Repository**: https://github.com/sauldataman/evm-signer
- **Python Client**: `skill/evm-signer/scripts/send_transaction.py`
- **Rule Schema**: `skill/evm-signer/references/rule_schema.md`
- **Example Config**: `conf/config.yaml.example`
