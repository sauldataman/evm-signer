# EVM Signer

> A secure transaction signing service for EVM-compatible blockchains
>
> Requirements: Go >= 1.19

## Getting Started

### Create Configuration Files

```shell
$ cp conf/config.yaml.example conf/config.yaml
$ cp conf/rule.json.example conf/rule.json
$ mkdir .keystore
```

### Configure config.yaml

#### Required Fields

```markdown
1. listen.port - The port number for the signing service
2. auth.ip - IP whitelist (e.g., 127.0.0.1)
3. account
    * Supported account types: Keystore, EvMnemonic, EncryptedMnemonic,
      PlainMnemonic, PlainPrivateKey
    * PlainMnemonic and PlainPrivateKey are recommended for testing only
    * Keystore: An encrypted JSON file containing a single private key
    * EvMnemonic: A virtual mnemonic combining multiple account types
    * EncryptedMnemonic: An encrypted mnemonic phrase file
    * Note: The "type" field is case-sensitive
```

### Rule Configuration

Use a JSON-formatted rule file for transaction validation. See `conf/rule.json.example` for reference.

#### JSON Rule Validation

The `conditions` array defines matching criteria. All conditions must be satisfied (AND logic) for a rule to match.

```json
[
  {
    "name": "example rule",
    "chain_id": 1,
    "conditions": [
      {
        "field": "to",
        "symbol": "==",
        "value": "0x1234567890abcdef1234567890abcdef12345678"
      }
    ]
  }
]
```

### Account Types

```markdown
1. PlainMnemonic
   A plaintext mnemonic phrase (supports 12, 15, 18, 21, or 24 words)

2. PlainPrivateKey
   A plaintext private key (66 characters with 0x prefix)

3. Keystore
   An encrypted JSON keystore file. Specify the file path in the "key" field.

4. EncryptedMnemonic
   An encrypted mnemonic file using the keystore algorithm.
   Specify the file path in the "key" field.

5. EvMnemonic
   A virtual mnemonic that combines multiple account types,
   exposing them as a single mnemonic-style account.
```

#### Account Configuration Examples

##### PlainMnemonic

```yaml
type: PlainMnemonic
key: <your_mnemonic>  # The mnemonic phrase
index: "0"            # Account index to use
```

##### PlainPrivateKey

```yaml
type: PlainPrivateKey
key: <your_private_key>  # The private key
```

##### EncryptedMnemonic

```yaml
type: EncryptedMnemonic
key: <path_to_encrypted_file>  # Path to the encrypted mnemonic file
pass: <password>               # Decryption password (optional; will prompt if omitted)
index: 0-10,11,20              # Account indices: ranges (0-10) or individual (11,20)
```

##### Keystore

```yaml
type: Keystore
key: <path_to_keystore>  # Path to the keystore file
pass: <password>         # Decryption password (optional; will prompt if omitted)
```

##### EvMnemonic

You can define multiple keys of different types. The `use_last_pass` option allows password reuse across sequential keys (parsed in ascending order by key number).

When `use_last_pass` is not set, it defaults to `false`. If `pass` is provided and `use_last_pass` is `true`, the explicit `pass` value takes priority.

```yaml
type: EvMnemonic
keys:
  1:
    type: Keystore
    key: ./account1.json
    pass: "******"
    use_last_pass: true
  2:
    type: PlainPrivateKey
    key: <your_private_key>
  3:
    type: PlainMnemonic
    key: "word1 word2 word3 ... word12"
    index: "0"
```

### Configure rule.json

Define signing rules based on transaction properties. Available fields for validation:

```markdown
1. to
   The destination address (contract address for contract calls)

2. data_selector
   The function selector (first 4 bytes of the data field), e.g., "0x12345678"

3. EIP-712 fields
   eip712.domain.name / eip712.domain.version / eip712.domain.chainId /
   eip712.domain.verifyingContract
   (See the EIP-712 specification for details)
```

## Build

```shell
$ go build -o signer
```

## Running the Service

```shell
# Load rules from conf directory (same search path as config.yaml)
# Default: conf/rule.json

./signer start --port 8080

# Specify a different rule file
./signer start --port 8080 --rule rule-prod.json
```

## Exposing the Signer to the Internet

### Using ngrok

1. Sign up at https://ngrok.com/
2. Download the ngrok client from the dashboard
3. Get your authtoken from "Getting Started" > "Your Authtoken"
4. Configure ngrok following the official documentation
5. Start ngrok: `./ngrok http 8080` (replace 8080 with your signer port)

## FAQ

### How do I generate a new key pair?

Use the `key generate` subcommand:

```shell
./signer key generate
```

### What tool should I use to create an encrypted keystore file?

Use any tool that supports Ethereum keystore encryption (e.g., geth, ethers.js, web3.js).

### What tool should I use to create an encrypted mnemonic file?

Use any tool that supports keystore-style encryption for mnemonic phrases.

### How do I verify that an account loaded successfully?

On startup, the signer prints the first address for each configured account. If you see the address in the logs, the account loaded successfully.

## Security Best Practices

1. **Cloud servers**: Configure SSH login whitelists and restrict access to the signer port. Delete configurations and release the server when no longer needed. Use IPv4 format for whitelist addresses.

2. **ngrok tunnels**: When using ngrok, configure the IP whitelist with IPv6 format (e.g., `::1`).

3. **Disabling whitelist**: Remove the whitelist configuration if you don't want IP restrictions.

4. **Rule updates**: Add new rules whenever you need to support additional transaction types.
