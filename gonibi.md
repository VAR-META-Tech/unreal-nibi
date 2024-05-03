# Go API

# [Account](#account)

- [CreateAccount](#createaccount)
- [ImportAccountFromMnemoic](#importaccountfrommnemoic)
- [ImportAccountFromPrivateKey](#importaccountfromprivatekey)
- [DeleteAccount](#deleteaccount)

# [Queries](#queries)

- [GetPrivKeyFromMnemonic](#getprivkeyfrommnemonic)
- [GetAccountByKeyName](#getaccountbykeyname)
- [GetAccountByAddress](#getaccountbyaddress)
- [QueryWasmContract](#guerywasmcontract)
- [QueryTXHash](#querytxhash)

# [Transactions](#transactions)

- [TransferToken](#transfertoken)
- [ExecuteWasmContract](#executewasmcontract)

# [Other](#other)

- [Network](#network)

## Account

### CreateAccount

```go
phrase := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
isSuccess := CreateAccount("test_key", phrase, "pass")
```

### ImportAccountFromMnemoic

```go
phrase := GenerateRecoveryPhrase()
check := ImportAccountFromMnemoic(phrase, "test_key")
```

### ImportAccountFromPrivateKey

```go
phrase := GenerateRecoveryPhrase()
keyName := "TestKey"
privKey := GetPrivKeyFromMnemonic(phrase, keyName)
```

### DeleteAccount

```go
phrase := GenerateRecoveryPhrase()
CreateAccount("test_key", phrase, "pass")
check := DeleteAccount("test_key", "pass")
```

## Queries

### QueryAccount

Base Account Struct:

```go
// Define the BaseAccount struct in Go
type BaseAccount struct {
	Address       []byte
	Coins         *Coins
	PubKey        []byte
	AccountNumber uint64
	Sequence      uint64
}
```

Query Accoune Example:

```go
phrase := GenerateRecoveryPhrase()
CreateAccount("test_key", phrase, "pass")
addr := GetAddressFromKeyName("test_key")
acc, err := QueryAccount(addr)
```

### GetPrivKeyFromMnemonic

```go
phrase := GenerateRecoveryPhrase()
privKey := GetPrivKeyFromMnemonic(phrase, "test_key")
```

### GetAccountByKeyName

```go
phrase := GenerateRecoveryPhrase()
CreateAccount("test_key", phrase, "pass")
addr := GetAddressFromKeyName("test_key")
```

### GetAccountByAddress

```go
phrase := GenerateRecoveryPhrase()
CreateAccount("test_key", phrase, "pass")
addr := GetAddressFromKeyName("test_key")
acc := GetAccountByAddress(addr)
```

### QueryWasmContract

```go
contractAddr := "nibi1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3slkhcux"
queryMsg := "{\"owner_of\": {\"token_id\": \"unique-nft-15\", \"include_expired\": false}}"
result := QueryWasmContract(contractAddr, queryMsg)
```

### QueryTXHash

```go
txResult := QueryTXHash(txHash)
```

## Transactions

### TransferToken

```go
phrase := "napkin rigid magnet grass plastic spawn replace hobby tray eternal pupil olive pledge nasty animal base bitter climb guess analyst fat neglect zoo earn"
adminPhases := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
// create account from these phrases first
// then use the address to test the transfer token
check1 := CreateAccount("admin", adminPhases, "pass")
check2 := CreateAccount("test", phrase, "pass")

adminAddr := GetAddressFromKeyName("admin")
testAddr := GetAddressFromKeyName("test")
denom := "unibi"
amount := 75
result := TransferToken(adminAddr, testAddr, denom, amount)
```

### ExecuteWasmContract

```go

adminPhases := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"

susscess := CreateAccount("admin", adminPhases, "pass")

adminAddr := GetAddressFromKeyName("admin")

contractAddr := "nibi1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3slkhcux"
executeMsg := "{\"mint\": {\"token_id\": \"unique-nft-15\", \"owner\": \"nibi1zy7amen6h5e4whcta4ac656l0whsalzmnqrkc5\", \"token_uri\": \"https://metadata.com/nft1.json\"}}"

txHash := ExecuteWasmContract(adminAddr, contractAddr, executeMsg, "unibi", 75)

time.Sleep(3 * time.Second)
txResult := QueryTXHash(txHash)

time.Sleep(3 * time.Second)

queryMsg := "{\"owner_of\": {\"token_id\": \"unique-nft-15\", \"include_expired\": false}}"
result := QueryWasmContract(contractAddr, queryMsg)
```

# Other

### Network

```go
var (
	LocalNetworkInfo = NetworkInfo{
		GrpcEndpoint:      "localhost:9090",
		LcdEndpoint:       "http://localhost:1317",
		TmRpcEndpoint:     "http://localhost:26657",
		WebsocketEndpoint: "ws://localhost:26657/websocket",
		ChainID:           "nibiru-localnet-0",
	}
	TestNetworkInfo = NetworkInfo{
		GrpcEndpoint:      "grpc.testnet-1.nibiru.fi:443",
		LcdEndpoint:       "https://lcd.testnet-1.nibiru.fi",
		TmRpcEndpoint:     "https://rpc.testnet-1.nibiru.fi:443",
		WebsocketEndpoint: "wss://rpc.testnet-1.nibiru.fi/websocket",
		ChainID:           "nibiru-testnet-1",
	}
	MainNetworkInfo = NetworkInfo{
		GrpcEndpoint:      "grpc.nibiru.fi:443",
		LcdEndpoint:       "https://lcd.nibiru.fi",
		TmRpcEndpoint:     "https://rpc.nibiru.fi:443",
		WebsocketEndpoint: "wss://rpc.nibiru.fi/websocket",
		ChainID:           "cataclysm-1",
	}
)
```

You can use switchNetwork funciton to switch the network

```go
success := SwitchNetwork("test")
```
