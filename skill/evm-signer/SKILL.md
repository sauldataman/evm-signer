---
name: evm-signer
description: Send EVM transactions and manage risk control rules for a local signer service. Use when users want to (1) send native ETH transfers or ERC20 token transfers through the signer at localhost:8080, (2) create or modify risk control rules (rule.json) for transaction whitelisting based on recipient addresses and amounts. Only handles native transfers and ERC20 transfers, not general contract calls.
---

# EVM Signer

## Overview

Sign and broadcast native ETH transfers and ERC20 token transfers via a local signer service. Create risk control rules to whitelist allowed transactions.

## Task 1: Send Transactions

### Native ETH Transfer

```bash
python3 ~/.claude/skills/evm-signer/scripts/send_transaction.py \
  --chain-id CHAIN_ID \
  --from SENDER_ADDRESS \
  --to RECIPIENT_ADDRESS \
  --value AMOUNT_IN_WEI
```

### ERC20 Token Transfer

```bash
python3 ~/.claude/skills/evm-signer/scripts/send_transaction.py \
  --chain-id CHAIN_ID \
  --from SENDER_ADDRESS \
  --token TOKEN_CONTRACT_ADDRESS \
  --to RECIPIENT_ADDRESS \
  --amount TOKEN_AMOUNT
```

### Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `--chain-id` | Yes | Blockchain network ID (1=Ethereum, 10=Optimism, 137=Polygon, etc.) |
| `--from` | Yes | Sender wallet address |
| `--to` | Yes | Recipient address |
| `--value` | For native | Amount in wei (1 ETH = 10^18 wei) |
| `--token` | For ERC20 | Token contract address |
| `--amount` | For ERC20 | Token amount in smallest unit |
| `--nonce` | No | Override auto-fetched nonce |
| `--gas-limit` | No | Override estimated gas |
| `--gas-price` | No | Override current gas price |
| `--dry-run` | No | Sign only, don't broadcast |

### Supported Chains

Ethereum (1), Optimism (10), BSC (56), Polygon (137), Arbitrum (42161), Base (8453), Avalanche (43114), Fantom (250), Gnosis (100), zkSync (324), Linea (59144), Scroll (534352), Mantle (5000), Blast (81457), Sepolia (11155111), Holesky (17000)

## Task 2: Create Risk Control Rules

Generate rule.json entries to whitelist specific transaction patterns.

### Rule Structure

```json
[
  {
    "name": "descriptive_rule_name",
    "chain_id": 1,
    "conditions": [
      {"field": "to", "symbol": "==", "value": "0x..."},
      {"field": "value", "symbol": "<=", "value": "1000000000000000000"}
    ]
  }
]
```

### Quick Reference

**Fields:** `from`, `to`, `value`, `data_selector`, `data`, `data_param`

**Symbols:**
- `==` exact match (addresses)
- `>=`, `<=` comparison (amounts, works with `value` and `data_param`)
- `in` match list: `"0xaddr1,0xaddr2"`

**ERC20 selector:** `0xa9059cbb` (transfer)

**ERC20 amount limit:** Use `data_param` with ABI to limit token transfer amounts

### Common Examples

Allow native transfers to specific address:
```json
{"field": "to", "symbol": "==", "value": "0xRecipient"}
```

Allow transfers up to 1 ETH:
```json
{"field": "value", "symbol": "<=", "value": "1000000000000000000"}
```

Allow ERC20 transfers of specific token:
```json
[
  {"field": "to", "symbol": "==", "value": "0xTokenContract"},
  {"field": "data_selector", "symbol": "==", "value": "0xa9059cbb"}
]
```

Limit ERC20 transfer amount (e.g., max 10 USDT):
```json
{
  "field": "data_param",
  "symbol": "<=",
  "value": "10000000",
  "abi": "{\"name\":\"transfer\",\"type\":\"function\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}]}",
  "param": "value"
}
```

For detailed schema, see [references/rule_schema.md](references/rule_schema.md).

## Resources

### scripts/send_transaction.py

Assembles transaction, calls signer API at `http://localhost:8080/v1/sign/transaction`, broadcasts via public RPC.

### references/rule_schema.md

Complete documentation of rule fields, symbols, and patterns.
