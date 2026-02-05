# Rule Schema Reference

## Structure

```json
[
  {
    "name": "rule_name",
    "chain_id": 1,
    "conditions": [
      {
        "field": "to",
        "symbol": "==",
        "value": "0x..."
      }
    ]
  }
]
```

## Fields

| Field | Description | Example Values |
|-------|-------------|----------------|
| `from` | Sender address | `0x1234...` |
| `to` | Recipient address (native) or contract address (ERC20) | `0xabcd...` |
| `value` | Transfer amount in wei (native only) | `1000000000000000000` |
| `data_selector` | Function selector (first 4 bytes of calldata) | `0xa9059cbb` |
| `data` | Full calldata | `0xa9059cbb000...` |
| `data_param` | ABI-decoded parameter from calldata (requires `abi` and `param`) | See below |

## Symbols

| Symbol | Description | Applicable To |
|--------|-------------|---------------|
| `==` | Exact match (case insensitive) | All fields |
| `>=` | Greater than or equal | `value`, `data_param` (numeric) |
| `<=` | Less than or equal | `value`, `data_param` (numeric) |
| `in` | Match any in comma-separated list | `from`, `to`, `data_selector` |
| `contains` | Substring match | `data` |
| `regex` | Regular expression match | All string fields |

## data_param Field

The `data_param` field allows comparing ABI-decoded parameters from transaction calldata. This is useful for limiting ERC20 transfer amounts or other contract call parameters.

### Required Properties

| Property | Description |
|----------|-------------|
| `field` | Must be `"data_param"` |
| `abi` | JSON-encoded ABI of the function (escaped string) |
| `param` | Name of the parameter to compare |
| `symbol` | Comparison operator (`==`, `<=`, `>=`) |
| `value` | Value to compare against (decimal string for uint256) |

### ABI Format

The `abi` field must be a JSON-escaped function ABI string:

```json
"{\"name\":\"transfer\",\"type\":\"function\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\"}]}"
```

### Supported Parameter Types

| Type | Comparison | Example |
|------|------------|---------|
| `uint256` / `uint*` | Numeric (`==`, `<=`, `>=`) | Token amounts |
| `address` | String (`==`, `in`) | Recipient addresses |
| `string` | String (`==`, `contains`, `regex`) | String parameters |

## Common Patterns

### Native Transfer Rules

Allow transfers to specific address:
```json
{
  "name": "allow_to_treasury",
  "chain_id": 1,
  "conditions": [
    {"field": "to", "symbol": "==", "value": "0x742d35Cc6634C0532925a3b844Bc9e7595f8fE2E"}
  ]
}
```

Allow transfers up to 1 ETH:
```json
{
  "name": "small_transfers",
  "chain_id": 1,
  "conditions": [
    {"field": "value", "symbol": "<=", "value": "1000000000000000000"}
  ]
}
```

Allow transfers to whitelist:
```json
{
  "name": "whitelist_recipients",
  "chain_id": 1,
  "conditions": [
    {"field": "to", "symbol": "in", "value": "0xaddr1,0xaddr2,0xaddr3"}
  ]
}
```

### ERC20 Transfer Rules

Allow ERC20 transfers (any token):
```json
{
  "name": "allow_erc20_transfer",
  "chain_id": 1,
  "conditions": [
    {"field": "data_selector", "symbol": "==", "value": "0xa9059cbb"}
  ]
}
```

Allow transfers of specific token:
```json
{
  "name": "allow_usdc_transfer",
  "chain_id": 1,
  "conditions": [
    {"field": "to", "symbol": "==", "value": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"},
    {"field": "data_selector", "symbol": "==", "value": "0xa9059cbb"}
  ]
}
```

Limit USDT transfers to max 10 USDT (decimals=6, so 10 USDT = 10000000):
```json
{
  "name": "usdt_transfer_limit_10",
  "chain_id": 1,
  "conditions": [
    {"field": "to", "symbol": "==", "value": "0xdAC17F958D2ee523a2206206994597C13D831ec7"},
    {"field": "data_selector", "symbol": "==", "value": "0xa9059cbb"},
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

### Combined Rules

Allow native transfers up to 0.1 ETH to whitelist:
```json
{
  "name": "limited_whitelist_transfers",
  "chain_id": 1,
  "conditions": [
    {"field": "to", "symbol": "in", "value": "0xaddr1,0xaddr2"},
    {"field": "value", "symbol": "<=", "value": "100000000000000000"}
  ]
}
```

## Important Notes

1. All conditions in a rule must match (AND logic)
2. Multiple rules are evaluated in order; first match wins
3. Addresses are case-insensitive
4. `value` comparisons use decimal strings (wei)
5. `chain_id` must match the transaction's chain ID
