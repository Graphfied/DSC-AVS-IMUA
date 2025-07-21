# Hello AVS Guide 

# **Hello World AVS**

Welcome to the **Hello World AVS**. This project demonstrates the simplest functionality you can expect from an AVS, providing a concrete understanding of its basic components.

**GitHub Repository**:

- [Hello World AVS](https://github.com/imua-xyz/hello-world-avs)
- [IMUA CLI](https://github.com/imua-xyz/imuachain)

### **AVS User Flow Example**

1. **Request**: An AVS consumer requests the generation and signing of a "Hello World" message.
2. **Event Emission**: The `HelloWorld` contract receives this request and emits a `NewTaskCreated` event.
3. **Task Processing by Operators**:
    - All registered Operators who have staked or delegated assets to the AVS pick up the request.
    - Each Operator generates the requested message, hashes it, and signs the hash using their private key.
4. **Submission**: Operators submit their signed hashes back to the `HelloWorld` AVS contract.
5. **Validation**: If the Operator is registered and meets the minimum staking requirements, their submission is accepted.

# Steps

## Run the node

<aside>
💡  

Prerequistion: [Install dependency](https://docs.imua.xyz/validator-setup/compiling-binary-from-source) or download it directly below.

</aside>  

[imuad]()

```powershell
git clone [https://github.com/imua-xyz/imuachain.git](https://github.com/imua-xyz/imuachain.git)
git checkout develop
cd imuachain
make build
cp build/imuad /usr/bin/imuad
./local_node.sh
```  

Check Available Accounts  

```powershell
imuad keys list --home ~/.tmp-imuad
- address: im16tltge7d4yr0wtkr7ut6dwwnqgnwm2ge63djdp
  name: dev0
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"AopCZPMxuF3pvVk+LV4MNXrejD6HCjMR+8tkB7Y+F4/F"}'
  type: local
  ...
```  

## Register Operator

Fund the account if the balance of the dev0 account is 0, check it with `imuad q bank balances im16tltge7d4yr0wtkr7ut6dwwnqgnwm2ge63djdp`

```powershell
imuad tx operator register-operator --meta-info "Operator1" --from dev0 --commission-rate 0.5 --commission-max-rate 1 --commission-max-change-rate 1 --home ~/.tmp-imuad --keyring-backend test --fees 50000000000hua --chain-id imuachainlocalnet_232-1
```  

## Show Operator ECDSA Key

```powershell
imuad keys unsafe-export-eth-key dev0 --home ~/.tmp-imuad
DBDE1905049DEE771ED652F3DC05D57EBFEC0FE1B0ED0310482E9C58DB398834
```  

## Import Keys

```powershell
git clone https://github.com/imua-xyz/hello-world-avs.git
cd hello-world-avs
make build
# avs owner private key sample: D196DCA836F8AC2FFF45B3C9F0113825CCBB33FA1B39737B948503B263ED75AE
# operator private key sample: DBDE1905049DEE771ED652F3DC05D57EBFEC0FE1B0ED0310482E9C58DB398834
# bls private key sample: 1c0599ffc52d512fd5b549fa050833e7d3bba12969d09d70a16441384e5a8a3a
./imua-key importKey --key-type ecdsa --private-key {avsOwner_private_key} --output-dir tests/keys/avs.ecdsa.key.json
./imua-key importKey --key-type ecdsa --private-key {operator_private_key} --output-dir tests/keys/operator.ecdsa.key.json
./imua-key  importKey --key-type bls --private-key {bls_private_key}  --output-dir tests/keys/test.bls.key.json  

```

Continue to configure the `config.yaml` file in `hello-world-avs/config.yaml`

- **operator_address**
    - It is the EIP-55 address of the bench32 address of the operator, check it with

    ```powershell
    imuad debug addr im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
    Address bytes: [175 8 51 107 198 168 209 105 195 74 78 158 77 98 118 227 98 244 117 126]
	Address (hex): AF08336BC6A8D169C34A4E9E4D6276E362F4757E
	Address (EIP-55): 0xAF08336BC6A8D169c34A4e9e4d6276e362F4757E
	Bech32 Acc: im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
    Bech32 Val: imvaloper14uyrx67x4rgkns62f60y6cnkud30gat769xuz5
    ```  

- **avs_owner_addresses**
    - The bench32 address of the operator must be included in **avs_owner_addresses**

```powershell
# this sets the logger level (true = info, false = debug)
production: false
#The eoa address of Operato, used for sending transactions
operator_address: 0x3e108c058e8066DA635321Dc3018294cA82ddEdf
#The eoa address of avs owner, used for sending transactions
avs_owner_address: 0x4b99E597121C99ba5846c32bd49d8A4B95457f8C
#Address of deployed AVS contract
avs_address: 0x10Ed22D975453A5D4031440D51624552E4f204D5
# ETH RPC URL
eth_rpc_url: http://127.0.0.1:8545
eth_ws_url: ws://127.0.0.1:8546
avs_ecdsa_private_key_store_path: tests/keys/avs.ecdsa.key.json
operator_ecdsa_private_key_store_path: tests/keys/operator.ecdsa.key.json
bls_private_key_store_path: tests/keys/test.bls.key.json
node_api_ip_port_address: 0.0.0.0:9010
enable_node_api: false
register_operator_on_startup: true
#register avs parameters
avs_name: "hello-avs"
min_stake_amount: 1
avs_owner_addresses:
  - "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C"
  - "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"
whitelist_addresses:
  - "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C"
  - "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"
asset_ids:
  - "0xdac17f958d2ee523a2206206994597c13d831ec7_0x65"
avs_unbonding_period: 7
min_self_delegation: 0
epoch_identifier: minute
avs_reward_address: 0x10Ed22D975453A5D4031440D51624552E4f204D5
avs_slash_address: 0x10Ed22D975453A5D4031440D51624552E4f204D5
task_address: 0x10Ed22D975453A5D4031440D51624552E4f204D5
mini_opt_in_operators: 3
min_total_stake_amount: 3
avs_reward_proportion: 3
avs_slash_proportion: 3
#create new task parameters
#Create task intervals,Unit second
create_task_interval: 100
task_response_period: 3
task_challenge_period: 3
threshold_percentage: 100
task_statistical_period: 3
# depoist and delegate params
deposit_amount: 100
delegate_amount: 100
staker: 0xa53f68563D22EB0dAFAA871b6C08a6852f91d627
```

## Register AVS and Create Task

```powershell
git clone https://github.com/imua-xyz/hello-world-avs.git
cd hello-world-avs
make build
./avsbinary/main --config config.yaml
```  

Result:

```powershell
./avsbinary/main --config config.yaml 
2024-11-13T12:27:10.955Z        INFO    logging/zap_logger.go:49        AVS_ADDRESS env var not set. will deploy avs contract
2024-11-13T12:27:11.644Z        INFO    logging/zap_logger.go:69        tx hash: 0x81477902f1c3432b933ef58a12f9e9990dd7d3913253af687ee412927a6f8f48
2024-11-13T12:27:11.644Z        INFO    logging/zap_logger.go:69        contract address: 0xd83404Cde3A28b751a661521Cb0aD3Cc35B7fa95
2024-11-13T12:27:17.352Z        DEBUG   logging/zap_logger.go:45        Estimating gas and nonce        {"tx": "0x08aa5b66c6f344519670509138e868dbe9e3cfef4bd275a6eb10430db11571b0"}
2024-11-13T12:27:17.365Z        DEBUG   logging/zap_logger.go:45        Getting signer for tx   {"tx": "0xc14fa15c9308af5b9853f776966896bd1236f606a094c125d6c56ebed1ce2abb"}
2024-11-13T12:27:18.064Z        DEBUG   logging/zap_logger.go:45        Sending transaction     {"tx": "0xc14fa15c9308af5b9853f776966896bd1236f606a094c125d6c56ebed1ce2abb"}
2024-11-13T12:27:20.072Z        INFO    logging/zap_logger.go:69        tx hash: 0x08aa5b66c6f344519670509138e868dbe9e3cfef4bd275a6eb10430db11571b0
2024-11-13T12:27:20.072Z        INFO    logging/zap_logger.go:69        Starting avs.
2024-11-13T12:27:20.072Z        INFO    logging/zap_logger.go:69        Avs owner set to send new task every 200 seconds
2024-11-13T12:27:20.072Z        INFO    logging/zap_logger.go:49        Avs sending new task
2024-11-13T12:27:20.074Z        WARN    logging/zap_logger.go:53        AVS USD value is zero or negative       {"value": "0.000000000000000000", "attempt": 1, "max_attempts": 10}
```  

It will show the contract address of the created AVS, and also the `config.yaml` will be overwritten with `avs_address`, `avs_reward_address` , `avs_slash_address`, `task_address`, Since there is no operator opt in to the AVS, so the AVS USD value is zero, it is not allowed to create new task.

## Operator Opt into AVS

```powershell
 imuad tx operator opt-into-avs <avs_contract_address> --from dev0 --home ~/.tmp-imuad --keyring-backend test --fees 50000000000hua --chain-id imuachainlocalnet_232-1
```  

Verify operator opt into the avs:

```powershell
imuad q operator get-avs-list im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
avs_list:
- 0xd83404cde3a28b751a661521cb0ad3cc35b7fa95
```  

## Deposit and Make Delegation

Prerequisite:

```powershell
git clone https://github.com/imua-xyz/hello-world-avs.git
cd hello-world-avs
make build
```  

Deposit and delegate with operator module in hello-world-avs

```powershell
./operatorbinary/main --config config.yaml
```  

Result:

```powershell
2024-11-13T12:34:44.882Z        INFO    logging/zap_logger.go:49        current epoch  is :     {"currentEpoch": 2837}
2024-11-13T12:34:44.882Z        INFO    logging/zap_logger.go:49        Execute Phase One Submission Task       {"currentEpoch": 2837, "startingEpoch": 2836, "taskResponsePeriod": 2}
2024-11-13T12:34:44.882Z        INFO    logging/zap_logger.go:49        Submitting task response for task response period       {"taskAddr": "0xd83404Cde3A28b751a661521Cb0aD3Cc35B7fa95", "taskId": 2, "operator-addr": "0xAF08336BC6A8D169c34A4e9e4d6276e362F4757E"}
2024-11-13T12:34:45.811Z        DEBUG   logging/zap_logger.go:45        Estimating gas and nonce        {"tx": "0x44c374ddf58b24aef97d1845eda88ab9478ac610dc4194275bfb6bfb255b6491"}
2024-11-13T12:34:45.821Z        DEBUG   logging/zap_logger.go:45        Getting signer for tx   {"tx": "0x2fbc117007fb43faa196ad4855bdb68cbfe97148587ebaf967f18bda9796af9c"}
2024-11-13T12:34:45.968Z        INFO    logging/zap_logger.go:69        tx hash: 0x94dbf2ce7c41fbe01abf8977f8623629e357f35eee727839156dd487f4bfb35a
2024-11-13T12:34:45.971Z        INFO    logging/zap_logger.go:49        current epoch  is :     {"currentEpoch": 2837}
2024-11-13T12:34:45.971Z        INFO    logging/zap_logger.go:49        Execute Phase Two Submission Task       {"currentEpoch": 2837, "startingEpoch": 2833, "taskResponsePeriod": 2, "taskStatisticalPeriod": 2}
2024-11-13T12:34:45.971Z        INFO    logging/zap_logger.go:49        Submitting task response for statistical period {"taskAddr": "0xd83404Cde3A28b751a661521Cb0aD3Cc35B7fa95", "taskId": 1, "operator-addr": "0xAF08336BC6A8D169c34A4e9e4d6276e362F4757E"}
```  

check the result of `./avs/main --config config.yaml` again.

```powershell
2024-11-13T12:30:09.594Z        INFO    logging/zap_logger.go:69        Starting avs.
2024-11-13T12:30:09.594Z        INFO    logging/zap_logger.go:69        Avs owner set to send new task every 200 seconds
2024-11-13T12:30:09.594Z        INFO    logging/zap_logger.go:49        Avs sending new task
2024-11-13T12:30:10.305Z        DEBUG   logging/zap_logger.go:45        Estimating gas and nonce        {"tx": "0xb273fc6702500277b74236b18d92c5ee529e7f949d4b794645059a2afa99400a"}
2024-11-13T12:30:10.320Z        DEBUG   logging/zap_logger.go:45        Getting signer for tx   {"tx": "0x1ec222369d5459b6e92b094da6d4feacb362b3635022cf51b8c20eb20f801fbd"}
2024-11-13T12:30:11.057Z        DEBUG   logging/zap_logger.go:45        Sending transaction     {"tx": "0x1ec222369d5459b6e92b094da6d4feacb362b3635022cf51b8c20eb20f801fbd"}
2024-11-13T12:30:13.061Z        INFO    logging/zap_logger.go:49        Transaction not yet mined       {"hash": "0x871859c4beda7559ae434e6c5506e16b43e2cd6e20e77b9d34a7a36bba6fb9d3"}
2024-11-13T12:30:15.063Z        INFO    logging/zap_logger.go:69        tx hash: 0xb273fc6702500277b74236b18d92c5ee529e7f949d4b794645059a2afa99400a
2024-11-13T12:33:29.593Z        INFO    logging/zap_logger.go:49        sendNewTask-num:        {"taskNum": 2}
2024-11-13T12:33:29.593Z        INFO    logging/zap_logger.go:49        Avs sending new task
2024-11-13T12:33:30.529Z        DEBUG   logging/zap_logger.go:45        Estimating gas and nonce        {"tx": "0x212ec28c1ee4eae4733d2d2274c4e036c814e40a1ae4f287e566cdccc347296d"}
2024-11-13T12:33:30.543Z        DEBUG   logging/zap_logger.go:45        Getting signer for tx   {"tx": "0xf50abe8ff4082bdf5b5d7aab1730e7c6aab2720e5ae18686742deda559491706"}
2024-11-13T12:33:31.395Z        DEBUG   logging/zap_logger.go:45        Sending transaction     {"tx": "0xf50abe8ff4082bdf5b5d7aab1730e7c6aab2720e5ae18686742deda559491706"}
2024-11-13T12:33:33.401Z        INFO    logging/zap_logger.go:69        tx hash: 0x212ec28c1ee4eae4733d2d2274c4e036c814e40a1ae4f287e566cdccc347296d
2024-11-13T12:36:49.591Z        INFO    logging/zap_logger.go:49        sendNewTask-num:        {"taskNum": 3}
```  

## Check TaskInfo with chain

```powershell
imuad q avs TaskInfo 0xd83404Cde3A28b751a661521Cb0aD3Cc35B7fa95 1
actual_threshold: "7766279631452241920"
err_signed_operators: []
hash: bN9XyPiULWhVmBM38FXdvn9gE46xQe91Sewfs+LAkXs=
name: Z1Ujo
no_signed_operators: []
operator_active_power:
  operator_power_list:
  - active_power: "315786.000000000000000000"
    operator_addr: im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
opt_in_operators:
- im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
signed_operators:
- im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
starting_epoch: "2833"
task_challenge_period: "2"
task_contract_address: 0xd83404cde3a28b751a661521cb0ad3cc35b7fa95
task_id: "1"
task_response_period: "2"
task_statistical_period: "2"
task_total_power: "315786.000000000000000000"
threshold_percentage: "100"
```  

## Check Task Submit Info with chain

If the phase is `PHASE_DO_COMMIT`, it is the expected result that two phase submit result is completed.

```powershell
imuad q avs SubmitTaskResult 0xd83404Cde3A28b751a661521Cb0aD3Cc35B7fa95 1 im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
info:
  bls_signature: jncXW+w8ZHcVnU/gK3F2GUktj0ZEdNxPLost64TN9Pl/KDBhl04ae/3PTxeHZ36LEU10IW2+/wdJO7njDnPUZFf9MokjbhgoTHlWsyCpJLoKBpadpBrAWcsYid856TwE
  operator_address: im14uyrx67x4rgkns62f60y6cnkud30gat7d5v30r
  phase: PHASE_DO_COMMIT
  task_contract_address: 0xd83404Cde3A28b751a661521Cb0aD3Cc35B7fa95
  task_id: "1"
  task_response: eyJUYXNrSUQiOjEsIk51bWJlclN1bSI6MX0=
  task_response_hash: 0x91b8eabcc462c7ded2c0427c3bbd3e4b05c5a73ea55571ff114b43dd9aeff767
```  
## Challenge
The task challenges are divided into two types: manually executing and automatically listening for events

### manually executing
```powershell
./challengebinary --config config.yaml --ExecType 2    --task-ID  1   --NumberToBeSquared 402
Note:Parameters must be filled in.
   --task-ID value            task ID (default: 0)
   --NumberToBeSquared value  number to be squared (default: 0)
Parameter sources:It is obtained from the log of the operatorbinary process window, which is obtained by parsing event
Examples：Received new task	{"id": "2", "name": "tYNvo", "numberToBeSquared": 270}
```  
### automatically executing
```powershell
./challengebinary --config config.yaml --ExecType 1
```  

Result:

```powershell
2025-02-19T14:22:48.151+0800	INFO	logging/zap_logger.go:69	The current task 0x10Ed22D975453A5D4031440D51624552E4f204D5--1 has been challenged:
2025-02-19T14:23:33.037+0800	INFO	logging/zap_logger.go:49	parse logs	{"data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAABLmeWXEhyZulhGwyvUnYpLlUV/jAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQ4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAFdFlOdm8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "height": 43, "event": [{"Name":"taskId","Type":{"Elem":null,"Size":256,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"issuer","Type":{"Elem":null,"Size":20,"T":7,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"name","Type":{"Elem":null,"Size":0,"T":3,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"numberToBeSquared","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"taskResponsePeriod","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"taskChallengePeriod","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"thresholdPercentage","Type":{"Elem":null,"Size":8,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"taskStatisticalPeriod","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false}]}
2025-02-19T14:23:33.037+0800	DEBUG	logging/zap_logger.go:45	Received new task	{"task": [2,"0x4b99e597121c99ba5846c32bd49d8a4b95457f8c","tYNvo",270,3,3,100,3]}
2025-02-19T14:23:33.037+0800	INFO	logging/zap_logger.go:49	Received new task	{"id": "2", "name": "tYNvo", "numberToBeSquared": 270}
2025-02-19T14:23:33.040+0800	INFO	logging/zap_logger.go:49	TriggerChallege	{"taskInfo": {"TaskContractAddress":"0x10ed22d975453a5d4031440d51624552e4f204d5","Name":"tYNvo","Hash":"SaenpZnYDurpJ7pFp6qa8w3au3+oOJhDsDnotXYf8Jk=","TaskID":2,"TaskResponsePeriod":3,"TaskStatisticalPeriod":3,"TaskChallengePeriod":3,"ThresholdPercentage":100,"StartingEpoch":44,"ActualThreshold":"","OptInOperators":["0x3e108c058e8066da635321dc3018294ca82ddedf"],"SignedOperators":[],"NoSignedOperators":[],"ErrSignedOperators":[],"TaskTotalPower":"0.000000000000000000","OperatorActivePower":[],"IsExpected":false,"EligibleRewardOperators":[],"EligibleSlashOperators":[]}}
2025-02-19T14:23:57.244+0800	INFO	logging/zap_logger.go:49	latest-taskInfo	{"taskInfo": {"TaskContractAddress":"0x10ed22d975453a5d4031440d51624552e4f204d5","Name":"tYNvo","Hash":"SaenpZnYDurpJ7pFp6qa8w3au3+oOJhDsDnotXYf8Jk=","TaskID":2,"TaskResponsePeriod":3,"TaskStatisticalPeriod":3,"TaskChallengePeriod":3,"ThresholdPercentage":100,"StartingEpoch":44,"ActualThreshold":"","OptInOperators":["0x3e108c058e8066da635321dc3018294ca82ddedf"],"SignedOperators":["0x3e108c058e8066da635321dc3018294ca82ddedf"],"NoSignedOperators":[],"ErrSignedOperators":[],"TaskTotalPower":"5100.000000000000000000","OperatorActivePower":[{"Operator":"0x0000000000000000000000000000000000000000","Power":5100000000000000000000}],"IsExpected":false,"EligibleRewardOperators":[],"EligibleSlashOperators":[]}}
2025-02-19T14:23:57.245+0800	INFO	logging/zap_logger.go:49	Execute raiseAndResolveChallenge	{"currentEpoch": 51, "startingEpoch": 44, "taskResponsePeriod": 3, "taskStatisticalPeriod": 3}
2025-02-19T14:23:57.905+0800	DEBUG	logging/zap_logger.go:45	Estimating gas and nonce	{"tx": "0x8a3ae07a4bb1955768d2afb936f51caf3fd3c3df23af5656d7b9dd72ae5c68fd"}
2025-02-19T14:23:57.913+0800	DEBUG	logging/zap_logger.go:45	Getting signer for tx	{"tx": "0x9adb229339307988f717a10f07d7f85b3557563cef1eb502d51ab1fb080a12d4"}
2025-02-19T14:23:58.576+0800	DEBUG	logging/zap_logger.go:45	Sending transaction	{"tx": "0x9adb229339307988f717a10f07d7f85b3557563cef1eb502d51ab1fb080a12d4"}
2025-02-19T14:24:00.265+0800	INFO	logging/zap_logger.go:49	parse logs	{"data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAQ7SLZdUU6XUAxRA1RYkVS5PIE1Q==", "height": 52, "event": [{"Name":"taskId","Type":{"Elem":null,"Size":256,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"issuer","Type":{"Elem":null,"Size":20,"T":7,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"name","Type":{"Elem":null,"Size":0,"T":3,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"numberToBeSquared","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"taskResponsePeriod","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"taskChallengePeriod","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"thresholdPercentage","Type":{"Elem":null,"Size":8,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false},{"Name":"taskStatisticalPeriod","Type":{"Elem":null,"Size":64,"T":1,"TupleRawName":"","TupleElems":null,"TupleRawNames":null,"TupleType":null},"Indexed":false}]}
2025-02-19T14:24:00.581+0800	INFO	logging/zap_logger.go:69	tx hash: 0x8a3ae07a4bb1955768d2afb936f51caf3fd3c3df23af5656d7b9dd72ae5c68fd
2025-02-19T14:24:00.581+0800	INFO	logging/zap_logger.go:69	The current task 0x10Ed22D975453A5D4031440D51624552E4f204D5--2 has been challenged:
```  
## Check ChallengeInfo with chain

```powershell
imuad q avs  ChallengeInfo 0x10Ed22D975453A5D4031440D51624552E4f204D5  1
challenge_address: 4b99e597121c99ba5846c32bd49d8a4b95457f8c
```  

## hello-cli
### prints operator status as viewed from avs contracts
```powershell
hello-cli --config config.yaml  print-operator-status

2025/02/20 01:15:54 Config: {
  "Production": false,
  "AVSOwnerAddress": "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C",
  "OperatorAddress": "0x3e108c058e8066DA635321Dc3018294cA82ddEdf",
  "AVSAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "EthRpcUrl": "http://127.0.0.1:8545",
  "EthWsUrl": "ws://localhost:8546",
  "BlsPrivateKeyStorePath": "tests/keys/test.bls.key.json",
  "OperatorEcdsaPrivateKeyStorePath": "tests/keys/operator.ecdsa.key.json",
  "AVSEcdsaPrivateKeyStorePath": "tests/keys/avs.ecdsa.key.json",
  "RegisterOperatorOnStartup": false,
  "NodeApiIpPortAddress": "0.0.0.0:9010",
  "EnableNodeApi": false,
  "AvsName": "hello-avs",
  "MinStakeAmount": 1,
  "AvsOwnerAddresses": [
    "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C",
    "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"
  ],
  "WhitelistAddresses": [
    "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C",
    "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"
  ],
  "AssetIDs": [
    "0xdac17f958d2ee523a2206206994597c13d831ec7_0x65"
  ],
  "AvsUnbondingPeriod": 7,
  "MinSelfDelegation": 0,
  "EpochIdentifier": "minute",
  "TaskAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "AVSRewardAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "AVSSlashAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "CreateTaskInterval": 50,
  "TaskResponsePeriod": 3,
  "TaskChallengePeriod": 3,
  "ThresholdPercentage": 100,
  "TaskStatisticalPeriod": 3,
  "MiniOptInOperators": 3,
  "MinTotalStakeAmount": 3,
  "AvsRewardProportion": 3,
  "AvsSlashProportion": 3,
  "DepositAmount": 100,
  "DelegateAmount": 100,
  "Staker": "0xa53f68563D22EB0dAFAA871b6C08a6852f91d627"
}
2025-02-20T01:15:54.911+0800    INFO    logging/zap_logger.go:49        OPERATOR_BLS_KEY_PASSWORD env var not set. using empty string
2025-02-20T01:15:55.518+0800    INFO    logging/zap_logger.go:49        OPERATOR_ECDSA_KEY_PASSWORD env var not set. using empty string
2025-02-20T01:15:55.519+0800    INFO    logging/zap_logger.go:49        operatorSender: {"operatorSender": "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"}
2025-02-20T01:16:00.525+0800    INFO    logging/zap_logger.go:49        Operator info   {"operatorAddr": "0x3e108c058e8066DA635321Dc3018294cA82ddEdf", "operatorKey": "trd13VAjRPoLHT8A/Umcw/wuQhG9frusRV7ZCh5ONPJxGyP9O+inGd/bY/TSWgXb"}
2025-02-20T01:16:00.525+0800    INFO    logging/zap_logger.go:49        Printing operator status
2025-02-20T01:16:00.525+0800    INFO    logging/zap_logger.go:49        {
 "EcdsaAddress": "0x3e108c058e8066DA635321Dc3018294cA82ddEdf",
 "PubkeysRegistered": true,
 "Pubkey": "\ufffd\ufffdu\ufffdP#D\ufffd\u000b\u001d?\u0000\ufffdI\ufffd\ufffd\ufffd.B\u0011\ufffd~\ufffd\ufffdE^\ufffd\n\u001eN4\ufffdq\u001b#\ufffd;\ufffd\ufffd\u0019\ufffd\ufffdc\ufffd\ufffdZ\u0005\ufffd",
 "RegisteredWithAvs": false

```  
### monitor
Subscribe to events using websocket,Monitor create and challenge tasks

```powershell
hello-cli --config config.yaml  monitor 
2025/02/20 01:25:38 Config: {
  "Production": false,
  "AVSOwnerAddress": "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C",
  "OperatorAddress": "0x3e108c058e8066DA635321Dc3018294cA82ddEdf",
  "AVSAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "EthRpcUrl": "http://127.0.0.1:8545",
  "EthWsUrl": "ws://localhost:8546",
  "BlsPrivateKeyStorePath": "tests/keys/test.bls.key.json",
  "OperatorEcdsaPrivateKeyStorePath": "tests/keys/operator.ecdsa.key.json",
  "AVSEcdsaPrivateKeyStorePath": "tests/keys/avs.ecdsa.key.json",
  "RegisterOperatorOnStartup": false,
  "NodeApiIpPortAddress": "0.0.0.0:9010",
  "EnableNodeApi": false,
  "AvsName": "hello-avs",
  "MinStakeAmount": 1,
  "AvsOwnerAddresses": [
    "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C",
    "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"
  ],
  "WhitelistAddresses": [
    "0x4b99E597121C99ba5846c32bd49d8A4B95457f8C",
    "0x3e108c058e8066DA635321Dc3018294cA82ddEdf"
  ],
  "AssetIDs": [
    "0xdac17f958d2ee523a2206206994597c13d831ec7_0x65"
  ],
  "AvsUnbondingPeriod": 7,
  "MinSelfDelegation": 0,
  "EpochIdentifier": "minute",
  "TaskAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "AVSRewardAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "AVSSlashAddress": "0x10Ed22D975453A5D4031440D51624552E4f204D5",
  "CreateTaskInterval": 50,
  "TaskResponsePeriod": 3,
  "TaskChallengePeriod": 3,
  "ThresholdPercentage": 100,
  "TaskStatisticalPeriod": 3,
  "MiniOptInOperators": 3,
  "MinTotalStakeAmount": 3,
  "AvsRewardProportion": 3,
  "AvsSlashProportion": 3,
  "DepositAmount": 100,
  "DelegateAmount": 100,
  "Staker": "0xa53f68563D22EB0dAFAA871b6C08a6852f91d627"
}
Starting event monitoring...
New Task Created:
  TaskID: 25
  Issuer: 0x4b99E597121C99ba5846c32bd49d8A4B95457f8C
  Name: R6LOT
  Number: 429
  Response Period: 3
  Challenge Period: 3
  Threshold: 100%
  Statistical Period: 3
Task Resolved:
  TaskID: 25
  Address: 0x10Ed22D975453A5D4031440D51624552E4f204D5

```  