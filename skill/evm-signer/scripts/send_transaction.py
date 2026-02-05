#!/usr/bin/env python3
"""
Send native ETH or ERC20 token transfers via EVM signer.

Usage:
    # Native ETH transfer
    python3 send_transaction.py --chain-id 1 --from 0x... --to 0x... --value 1000000000000000000

    # ERC20 transfer
    python3 send_transaction.py --chain-id 1 --from 0x... --token 0x... --to 0x... --amount 1000000

    # With custom nonce and gas
    python3 send_transaction.py --chain-id 1 --from 0x... --to 0x... --value 1000000000000000000 \
        --nonce 5 --gas-limit 21000 --gas-price 20000000000
"""

import argparse
import json
import sys
import urllib.request
import urllib.error

SIGNER_URL = "http://localhost:8080"

# Common public RPCs by chain ID
PUBLIC_RPCS = {
    1: "https://go.getblock.io/6db279c1e07c481da0785c453b4c5de1",
    10: "https://mainnet.optimism.io",
    56: "https://bsc-dataseed.binance.org",
    137: "https://polygon-rpc.com",
    42161: "https://arb1.arbitrum.io/rpc",
    8453: "https://mainnet.base.org",
    43114: "https://api.avax.network/ext/bc/C/rpc",
    250: "https://rpc.ftm.tools",
    100: "https://rpc.gnosischain.com",
    324: "https://mainnet.era.zksync.io",
    59144: "https://rpc.linea.build",
    534352: "https://rpc.scroll.io",
    5000: "https://rpc.mantle.xyz",
    81457: "https://rpc.blast.io",
    # Testnets
    11155111: "https://rpc.sepolia.org",
    17000: "https://ethereum-holesky.publicnode.com",
}

# ERC20 transfer function selector
ERC20_TRANSFER_SELECTOR = "0xa9059cbb"


def get_rpc_url(chain_id: int) -> str:
    if chain_id not in PUBLIC_RPCS:
        raise ValueError(f"Unknown chain ID: {chain_id}. Supported: {list(PUBLIC_RPCS.keys())}")
    return PUBLIC_RPCS[chain_id]


def rpc_call(rpc_url: str, method: str, params: list) -> dict:
    payload = json.dumps({
        "jsonrpc": "2.0",
        "method": method,
        "params": params,
        "id": 1
    }).encode()

    req = urllib.request.Request(
        rpc_url,
        data=payload,
        headers={"Content-Type": "application/json"}
    )

    with urllib.request.urlopen(req, timeout=30) as resp:
        result = json.loads(resp.read().decode())

    if "error" in result:
        raise Exception(f"RPC error: {result['error']}")
    return result.get("result")


def get_nonce(rpc_url: str, address: str) -> int:
    result = rpc_call(rpc_url, "eth_getTransactionCount", [address, "pending"])
    return int(result, 16)


def get_gas_price(rpc_url: str) -> int:
    result = rpc_call(rpc_url, "eth_gasPrice", [])
    return int(result, 16)


def estimate_gas(rpc_url: str, tx: dict) -> int:
    result = rpc_call(rpc_url, "eth_estimateGas", [tx])
    return int(result, 16)


def encode_erc20_transfer(to: str, amount: int) -> str:
    """Encode ERC20 transfer(address,uint256) call data."""
    to_padded = to.lower().replace("0x", "").zfill(64)
    amount_hex = hex(amount)[2:].zfill(64)
    return ERC20_TRANSFER_SELECTOR + to_padded + amount_hex


def sign_transaction(chain_id: int, tx: dict) -> str:
    """Call signer API to sign transaction.

    API expects:
    - POST /v1/sign/transaction
    - Content-Type: application/x-www-form-urlencoded
    - Form field 'data' containing JSON with:
      - chain_id: int
      - account: sender address
      - transaction: JSON string of transaction details
    """
    url = f"{SIGNER_URL}/v1/sign/transaction"

    # Build transaction object matching types.Transaction
    transaction = {
        "to": tx["to"],
        "value": tx.get("value", "0x0"),
        "gas": tx["gas"],
        "gasPrice": tx["gasPrice"],
        "nonce": tx["nonce"],
        "input": tx.get("data", "0x"),
    }

    # Build request matching types.MsgInfo
    sign_data = {
        "chain_id": chain_id,
        "account": tx["from"],
        "transaction": json.dumps(transaction),
    }

    form_data = f"data={urllib.parse.quote(json.dumps(sign_data))}"

    req = urllib.request.Request(
        url,
        data=form_data.encode(),
        headers={"Content-Type": "application/x-www-form-urlencoded"},
        method="POST"
    )

    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            result = json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        error_body = e.read().decode()
        try:
            error_json = json.loads(error_body)
            raise Exception(f"Signer error: {error_json.get('msg', error_body)}")
        except json.JSONDecodeError:
            raise Exception(f"Signer error ({e.code}): {error_body}")

    # Success response is Sign struct: {signature, tx, tx_hex}
    if "tx_hex" not in result:
        raise Exception(f"Unexpected response: {result}")

    return result["tx_hex"]


def broadcast_transaction(rpc_url: str, raw_tx: str) -> str:
    """Broadcast signed transaction to network."""
    result = rpc_call(rpc_url, "eth_sendRawTransaction", [raw_tx])
    return result


def main():
    parser = argparse.ArgumentParser(description="Send native ETH or ERC20 transfers")
    parser.add_argument("--chain-id", type=int, required=True, help="Chain ID")
    parser.add_argument("--from", dest="from_addr", required=True, help="Sender address")
    parser.add_argument("--to", required=True, help="Recipient address (for native) or token contract (for ERC20)")
    parser.add_argument("--value", type=int, default=0, help="Value in wei (for native transfer)")
    parser.add_argument("--token", help="ERC20 token contract address (if ERC20 transfer)")
    parser.add_argument("--amount", type=int, help="Token amount in smallest unit (for ERC20)")
    parser.add_argument("--nonce", type=int, help="Transaction nonce (auto if not specified)")
    parser.add_argument("--gas-limit", type=int, help="Gas limit (auto if not specified)")
    parser.add_argument("--gas-price", type=int, help="Gas price in wei (auto if not specified)")
    parser.add_argument("--dry-run", action="store_true", help="Only sign, don't broadcast")

    args = parser.parse_args()

    # Validate ERC20 transfer args
    if args.token and not args.amount:
        parser.error("--amount is required for ERC20 transfers")

    rpc_url = get_rpc_url(args.chain_id)
    print(f"Using RPC: {rpc_url}")

    # Determine if this is native or ERC20 transfer
    if args.token:
        # ERC20 transfer
        to_addr = args.token
        value = "0x0"
        data = encode_erc20_transfer(args.to, args.amount)
        print(f"ERC20 transfer: {args.amount} tokens to {args.to}")
    else:
        # Native transfer
        to_addr = args.to
        value = hex(args.value)
        data = "0x"
        print(f"Native transfer: {args.value} wei to {args.to}")

    # Get nonce
    nonce = args.nonce if args.nonce is not None else get_nonce(rpc_url, args.from_addr)
    print(f"Nonce: {nonce}")

    # Get gas price
    gas_price = args.gas_price if args.gas_price else get_gas_price(rpc_url)
    print(f"Gas price: {gas_price} wei")

    # Build transaction for gas estimation
    tx = {
        "from": args.from_addr,
        "to": to_addr,
        "value": value,
        "data": data,
        "nonce": hex(nonce),
        "gasPrice": hex(gas_price),
    }

    # Estimate gas
    if args.gas_limit:
        gas_limit = args.gas_limit
    else:
        gas_limit = estimate_gas(rpc_url, tx)
        gas_limit = int(gas_limit * 1.1)  # Add 10% buffer

    tx["gas"] = hex(gas_limit)
    print(f"Gas limit: {gas_limit}")

    # Sign transaction
    print("\nSigning transaction...")
    raw_tx = sign_transaction(args.chain_id, tx)
    print(f"Signed transaction: {raw_tx[:66]}...")

    if args.dry_run:
        print("\n[Dry run] Transaction signed but not broadcast")
        print(f"Full raw transaction: {raw_tx}")
        return

    # Broadcast transaction
    print("\nBroadcasting transaction...")
    tx_hash = broadcast_transaction(rpc_url, raw_tx)
    print(f"Transaction hash: {tx_hash}")


if __name__ == "__main__":
    import urllib.parse
    main()
