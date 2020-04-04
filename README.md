# iotex-hermes
The automatic reward distribution service for IoTeX delegates

## Get started

### Minimum requirements

| Components | Version | Description |
|----------|-------------|-------------|
| [Golang](https://golang.org) | &ge; 1.11.5 | Go programming language |

## Run as a service
1. If you put the project code under your `$GOPATH/src`, you will need to set up an environment variable:
```
export GO111MODULE=on
```

2. Specify distributor's account information by setting up the password for the keystore file and the IoTeX address for keystore verification:
```
export VAULT_PASSWORD=password_for_distributor's_keystore_file
export VAULT_ADDRESS=distributor's_IoTeX_address
```
Note that you need to set up a distributor account and generate the corresponding keystore file beforehand.

3. Specify IoTeX Public API address:
```
export IO_ENDPOINT=Full_Node_IP:API_Port
```
If the distrbution happens on IoTeX mainnet, you can use the MainNet secure endpoint:
```
api.iotex.one:443
```
If the distribution happens on IoTeX testnet, you can use the TestNet secure endpoint:
```
api.testnet.iotex.one:443
```

4. Specify smart contract addresses:
```
export MULTISEND_CONTRACT_ADDRESS=multisend_contract_address
export HERMES_CONTRACT_ADDRESS=hermes_contract_address
```
Note that it is **required** that you have deployed both contracts before starting the distribution service. 
Please refer to the [instruction](https://docs.iotex.io/#deploy-contract) if you want to know how to deploy a smart contract with ioctl command line tool.

5. You may need to distribute rewards in batches due to the gas limit constraint. To set how many distributions in a batch:
```
export CHUNK_SIZE=distribution_batch_size
```

6. Build service:
```
make build
```

7. Before distributing rewards, you may need to claim rewards first by executing the following command:
```
./bin/hermes claim DELEGATE
```

8. Distribute rewards to voters by executing the following command:
```
./bin/hermes distribute DELEGATE
```
