# Hello World AVS
# Overview

Welcome to the Hello World AVS. This project shows you the simplest functionality you can expect from an AVS. It will give you a concrete understanding of the basic components.

## Workflow
1. AVS consumer requests a "Hello World" message to be generated and signed.
2. HelloWorld contract receives the request and emits a NewTaskCreated event for the request.
3. All Operators who are registered to the AVS and has staked, delegated assets takes this request. Operator generates the requested message, hashes it, and signs the hash with their private key.
4. Each Operator submits their signed hash back to the HelloWorld AVS contract.
5. If the Operator is registered to the AVS and has the minimum needed stake, the submission is accepted.

# Installation

`make build`

Or check out the latest release.

# Quick Start

1.`./imua-key import --key-type ecdsa {pri_key}`   
2.`./avs/main --config config.yaml`  
3.`./operator/main --config config.yaml`  

## Private Key Management for AVS and Operator

### Overview
This document provides a comprehensive guide to generating and managing private keys for the AVS (Actively Validated Service) system using the `imua-key` utility.

### Prerequisites
- Ensure `imua-key` is installed(use "make imua-key" command)
- Have private keys ready for import

##  Steps

### 1. Import AVS ECDSA Private Key
```bash
# Import AVS private key
./imua-key importKey --key-type ecdsa --private-key {avs_private_key}  --output-dir tests/keys/avs.ecdsa.key.json
# Output: tests/keys/avs.ecdsa.key.json
```

### 2. Import Operator ECDSA Private Key
```bash
# Import Operator private key
./imua-key importKey --key-type ecdsa --private-key {operator_private_key} --output-dir tests/keys/operator.ecdsa.key.json
# Output: tests/keys/operator.ecdsa.key.json
```

### 3.Generate or Import BLS Private Key for operator
#### Import
```bash
# Import BLS private key
./imua-key  importKey --key-type bls --private-key {bls_private_key}  --output-dir tests/keys/test.bls.key.json
# Output: tests/keys/test.bls.key.json
```

#### Generate
```bash
# generate BLS key
./imua-key  generate --key-type bls --num-keys 1
# Output: random folder
```

### Key Paths
- AVS Owner ECDSA Private Key: `tests/keys/avs.ecdsa.key.json`
- Operator ECDSA Private Key: `tests/keys/operator.ecdsa.key.json`
- BLS Private Key: `tests/keys/test.bls.key.json`

## Key Management Best Practices
- Keep private keys secure
- Never share private keys
- Use environment-specific key management
- Rotate keys periodically

## Verification
```bash
# Verify key files exist
ls tests/keys/
```
## Note 🚨
**Important Password Configuration**

If a non-empty password is used when writing private keys to JSON files, you must set the corresponding environment variables:

| Key Type | Environment Variable | JSON File Path |
|----------|---------------------|----------------|
| BLS Key | `OPERATOR_BLS_KEY_PASSWORD` | `tests/keys/test.bls.key.json` |
| Operator ECDSA Key | `OPERATOR_ECDSA_KEY_PASSWORD` | `tests/keys/operator.ecdsa.key.json` |
| AVS ECDSA Key | `AVS_ECDSA_KEY_PASSWORD` | `tests/keys/avs.ecdsa.key.json` |

## Secure Environment Variable Management

### 1. Temporary Terminal Session
```bash
# Set individual key passwords
export OPERATOR_BLS_KEY_PASSWORD="strong_bls_password"
export OPERATOR_ECDSA_KEY_PASSWORD="strong_ecdsa_password"
export AVS_ECDSA_KEY_PASSWORD="strong_avs_password"
```

### 2. Persistent Zsh Configuration
```bash
# Edit ~/.zshrc configuration
nano ~/.zshrc

# Add secure passwords
export OPERATOR_BLS_KEY_PASSWORD="strong_bls_password"
export OPERATOR_ECDSA_KEY_PASSWORD="strong_ecdsa_password"
export AVS_ECDSA_KEY_PASSWORD="strong_avs_password"

# Reload configuration
source ~/.zshrc
```

### Configuration
#### config.yaml (path/config.yaml)
```
# this sets the logger level (true = info, false = debug)
production: false
#The eoa address of Operator, used for sending transactions
operator_address: 0xce5b680d1fd259ada4820e9314bcf0723bdb1287
#The eoa address of avs owner, used for sending transactions
avs_owner_address: 0x3e108c058e8066DA635321Dc3018294cA82ddEdf
#Address of deployed AVS contract
avs_address: 0xfF8f8297BEF982ac6ED7a203e144D9fa4F0FcE31
```

- **operator_address**
EIP-55 address for the operator, convert the address from bench32 address with imuad.

```
imuad debug addr <bench32_address>
```

for example
```
imuad debug addr im18cggcpvwspnd5c6ny8wrqxpffj5zmhkl3agtrj
Address bytes: [62 16 140 5 142 128 102 218 99 83 33 220 48 24 41 76 168 45 222 223]
Address (hex): 3E108C058E8066DA635321DC3018294CA82DDEDF
Address (EIP-55): 0x3e108c058e8066DA635321Dc3018294cA82ddEdf
Bech32 Acc: im18cggcpvwspnd5c6ny8wrqxpffj5zmhkl3agtrj
Bech32 Val: imvaloper18cggcpvwspnd5c6ny8wrqxpffj5zmhklxvzxw9
```

- **avs_owner_address**
AVS Owner address is used to deploy the avs contract on imua chain, it should be consistent with the avs_owner_addresses below. please note that avs_owner_addresses use bench32 address, while avs_owner_address use EIP-55 address.

- **avs_address**
The avs contract address deployed on imua chain. If it is empty string, it will be deployed on-fly.

```
# ETH RPC URL
eth_rpc_url: http://127.0.0.1:8545
eth_ws_url: ws://127.0.0.1:8546
avs_ecdsa_private_key_store_path: tests/keys/avs.ecdsa.key.json
operator_ecdsa_private_key_store_path: tests/keys/operator.ecdsa.key.json
bls_private_key_store_path: tests/keys/test.bls.key.json
node_api_ip_port_address: 0.0.0.0:9010
enable_node_api: true
register_operator_on_startup: false
```

- avs_ecdsa_private_key_store_path
- operator_ecdsa_private_key_store_path
- bls_private_key_store_path
After import avs owner/operator/bls keys with `imua-key` command, the json files will be generted under tests/keys folder.

```
#register avs parameters
avs_name: "hello-avs"
min_stake_amount: 1
avs_owner_addresses:
  - "im18cggcpvwspnd5c6ny8wrqxpffj5zmhkl3agtrj"
  - "im1eedksrgl6fv6mfyzp6f3f08swgaaky58xe5m7k"
asset_ids:
  - "0xdac17f958d2ee523a2206206994597c13d831ec7_0x65"
avs_unbonding_period: 7
min_self_delegation: 0
epoch_identifier: minute
avs_reward_address: 0x4f5CDaE0B1afeB0473dEb5AC4F9912409BBBBb72
avs_slash_address:  0x4f5CDaE0B1afeB0473dEb5AC4F9912409BBBBb72
task_address:  0x4f5CDaE0B1afeB0473dEb5AC4F9912409BBBBb72
params:
  - 5
  - 7
  - 8
  - 4
```
- **avs_name**
User specified AVS name.
- **min_stake_amount**
Minimal stake amount for the AVS.
- **avs_owner_addresses**
AVS Owner addresses that has access control for the AVS.
- **asset_ids**
Asset IDs that are supported by the AVS.
- **avs_unbonding_period**
Unbonding period for the AVS.
- **min_self_delegation**
Minimal self delegation for the AVS.
- **epoch_identifier**
Epoch identifier for the AVS. There are four types of epoch identifier: minute, hour, day, week.
- **avs_reward_address**
- **avs_slash_address**
- **task_address**
If avs_address is empty above, after the AVS contract is deployed on fly, the avs reward address, slash address and task address will be overwritten by the AVS contract address.
- **params**
1. minimal number of opt in operators.
2. minimal total stake amount.
3. reward percentage.
4. slash percentage.

```
#create new task parameters
#Create task intervals,Unit second
create_task_interval: 500
task_response_period: 2
task_challenge_period: 2
threshold_percentage: 100
task_statistical_period: 2
```

- **create_task_interval**
Create task interval(second), 500 stands for create a new task for every 500 seconds.
- **task_response_period**
Task response period(epoch),during epoch (the starting epoch , starting epoch + task_response_period] the operator is allowed to submit phase one result.
- **task_challenge_period**
Task challenge period(epoch), during epoch (the starting epoch + task_response_period + task_statistical_period , starting epoch + task_response_period + task_statistical_period + task_challenge_period], anyone can raise a challenge for a task.
- **threshold_percentage**
Threshold percentage for a task.
- **task_statistical_period**
Task statistical period(epoch), during epoch (the starting epoch + task_response_period, the starting epoch + task_response_period + task_statistical_period ], the operator is allowed to submiit phase two result.
