# Deploy cw721_base.wasm to Mint NFT in localnet

### Download cw721_base.wasm file follow this [link](https://github.com/CosmWasm/cosmwasm-plus/releases/download/v0.9.0/cw721_base.wasm):

```sh
cd <to folder store cw721_base.wasm file>
```

### Install nibid cli:

Follow this link to install nibid: https://nibiru.fi/docs/dev/cli/nibid-binary.html

### Setup your nibid config, make sure you run nibiru localnet

```sh
nibid config chain-id nibiru-localnet-0 && \
nibid config broadcast-mode sync && \
nibid config node "http://localhost:26657" && \
nibid config keyring-backend os && \
nibid config output json
```

## Deploy

```sh
FROM=nibi1zaavvzxez0elundtn32qnk9lkm8kmcsz44g7xl #validator address
```

```zsh

nibid tx wasm store cw721_base.wasm \
--from $FROM \
--gas auto \
--gas-adjustment 1.5 \
--gas-prices 0.025unibi \
--yes
```

```
TXHASH="$(nibid tx wasm store cw721_base.wasm \
--from $FROM \
--gas auto \
--gas-adjustment 1.5 \
--gas-prices 0.025unibi \
--yes | jq -rcs '.[0].txhash')"
```

```
nibid q tx $TXHASH > txhash.json
CODE_ID="$(cat txhash.json | jq -r '.logs[0].events[1].attributes[1].value')"
```

### Create a inst.json file:

`inst.json`

```json
{
  "name": "Vameta NFT",
  "symbol": "VNFT",
  "minter": "nibi1zaavvzxez0elundtn32qnk9lkm8kmcsz44g7xl"
}
```

```zsh
nibid tx wasm instantiate $CODE_ID \
"$(cat inst.json)" \
--admin "$FROM" \
--label Helo \
--from $FROM \
--gas auto \
--gas-adjustment 1.5 \
--gas-prices 0.025unibi \
--yes
```

```zsh
TXHASH_INIT="$(nibid tx wasm instantiate $CODE_ID \
"$(cat inst.json)" \
--admin "$FROM" \
--label Helo \
--from $FROM \
--gas auto \
--gas-adjustment 1.5 \
--gas-prices 0.025unibi \
--yes | jq -rcs '.[0].txhash')"
```

```zsh
nibid q tx $TXHASH_INIT > txhash.init.json
```

```
CONTRACT_ADDRESS="$(cat txhash.init.json | jq -r '.logs[0].events[1].attributes[0].value')"
```

## Once you have the CONTRACT_ADDRESS. You good to mint an NFT with this code example

```c
const adminAddress = "nibi1zaavvzxez0elundtn32qnk9lkm8kmcsz44g7xl"
char *testTx2 = ExecuteWasmContract(adminAddress, "nibi1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3slkhcux", // contract address
// msg to mint nft
"{\"mint\": {\"token_id\": \"unique-nft-15\", \"owner\": \"nibi1zy7amen6h5e4whcta4ac656l0whsalzmnqrkc5\", \"token_uri\": \"https://metadata.com/nft1.json\"}}",
"unibi", 1);

if (testTx2 == NULL)
{
	printf("Failed to Test ExecuteWasmContract\n");
	return 1;
}
else
{
	printf("TxHash %s\n", testTx2);
}
```
