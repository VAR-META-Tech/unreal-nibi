package test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/Unique-Divine/gonibi"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/sirupsen/logrus"
)

// Define the KeyInfo struct in Go
type KeyInfo struct {
	Type    uint32
	Name    string
	PubKey  []byte
	Address []byte
}

// Define the UserAccount struct in Go
type UserAccount struct {
	Info     *KeyInfo
	Password string
}

// Define the Coin struct in Go
type Coin struct {
	Denom  string
	Amount uint64
}

// Define the Coins struct in Go
type Coins struct {
	Array  []*Coin
	Length uint64
}

// Define the BaseAccount struct in Go
type BaseAccount struct {
	Address       []byte
	Coins         *Coins
	PubKey        []byte
	AccountNumber uint64
	Sequence      uint64
}

func main() {}

type NetworkInfo struct {
	GrpcEndpoint      string
	LcdEndpoint       string
	TmRpcEndpoint     string
	WebsocketEndpoint string
	ChainID           string
}

var (
	LocalNetworkInfo = NetworkInfo{
		GrpcEndpoint:      "localhost:9090",
		LcdEndpoint:       "http://localhost:1317",
		TmRpcEndpoint:     "http://localhost:26657",
		WebsocketEndpoint: "ws://localhost:26657/websocket",
		ChainID:           "nibiru-localnet-0",
	}
	DevNetworkInfo = NetworkInfo{
		GrpcEndpoint:      "tcp://grpc.devnet-2.nibiru.fi:443",
		LcdEndpoint:       "http://localhost:1317",
		TmRpcEndpoint:     "https://rpc.devnet-2.nibiru.fi:443",
		WebsocketEndpoint: "wss://rpc.devnet-2.nibiru.fi/websocket",
		ChainID:           "nibiru-devnet-2",
	}
	TestNetworkInfo = NetworkInfo{
		GrpcEndpoint:      "tcp://grpc.testnet-1.nibiru.fi:9090",
		LcdEndpoint:       "https://lcd.testnet-1.nibiru.fi",
		TmRpcEndpoint:     "https://rpc.testnet-1.nibiru.fi::443",
		WebsocketEndpoint: "wss://rpc.testnet-1.nibiru.fi/websocket",
		ChainID:           "nibiru-testnet-1",
	}
	MainNetworkInfo = NetworkInfo{
		GrpcEndpoint:      "localhost:9090",
		LcdEndpoint:       "http://localhost:1317",
		TmRpcEndpoint:     "https://rpc.nibiru.fi:443",
		WebsocketEndpoint: "ws://localhost:26657/websocket",
		ChainID:           "cataclysm-1",
	}
)

var gosdk gonibi.NibiruClient
var authClient authtypes.QueryClient
var bankClient banktypes.QueryClient
var wasmClient wasmtypes.QueryClient
var networkInfo NetworkInfo

func InitClients() error {
	authClient = authtypes.NewQueryClient(gosdk.Querier.ClientConn)
	bankClient = banktypes.NewQueryClient(gosdk.Querier.ClientConn)
	wasmClient = wasmtypes.NewQueryClient(gosdk.Querier.ClientConn)
	if authClient == nil || bankClient == nil || wasmClient == nil {
		return errors.New("can't init client")
	}
	return nil
}

func PrintPayload(funcName string, args ...interface{}) {
	// Log the function name
	logrus.WithField("Name", funcName).Info("\n\nCall function")

	// Log the function parameters
	for i, arg := range args {
		logrus.WithFields(logrus.Fields{"arg": i, "value": fmt.Sprintf("%v", arg)}).Debug("Parameter")
	}
}
func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	networkInfo = LocalNetworkInfo

	grpcConn, err := gonibi.GetGRPCConnection(networkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		logrus.Fatalf("Failed to initialize Nibiru client: %s", err)
	}

	gosdk, err = gonibi.NewNibiruClient(networkInfo.ChainID, grpcConn, networkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.Fatalf("Failed to initialize Nibiru client: %s", err)
	}

	if err := InitClients(); err != nil {
		logrus.Println("[ERR] ", err)
	}
	logrus.Println("[init] Nibiru client initialized")
}

//export SwitchNetwork
func SwitchNetwork(network string) int {
	logrus.Info("Call SwitchNetwork")
	networkStr := network
	switch networkStr {
	case "local":
		networkInfo = LocalNetworkInfo
	case "dev":
		networkInfo = DevNetworkInfo
	case "test":
		networkInfo = TestNetworkInfo
	case "main":
		networkInfo = MainNetworkInfo
	default:
		networkInfo = TestNetworkInfo
	}
	grpcConn, err := gonibi.GetGRPCConnection(networkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		logrus.Fatalf("Failed to initialize Nibiru client: %s", err)
		return Fail
	}

	gosdk, err = gonibi.NewNibiruClient(networkInfo.ChainID, grpcConn, networkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.Fatalf("Failed to initialize Nibiru client: %s", err)
		return Fail
	}

	logrus.Info("Switch to network: ", networkStr)

	if err := InitClients(); err != nil {
		logrus.Error("Init client err: ", err)
		return Fail
	}
	return Success
}

const (
	Success = 0
	Fail    = 1
)

// Niburu method

func GetAccountInfo(
	address string,
) (account authtypes.AccountI, err error) {

	acc, err := authClient.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}
	var accountI authtypes.AccountI
	gosdk.EncCfg.InterfaceRegistry.UnpackAny(acc.Account, &accountI)
	return accountI, nil
}

func GetListAccountInfo() (accouns []authtypes.AccountI, err error) {

	queryClient := authtypes.NewQueryClient(gosdk.Querier.ClientConn)
	resp, err := queryClient.Accounts(context.Background(), &authtypes.QueryAccountsRequest{})
	if err != nil {
		return accouns, err
	}
	// register auth interface
	var accounts []authtypes.AccountI
	for _, v := range resp.Accounts {
		var acc authtypes.AccountI
		gosdk.EncCfg.InterfaceRegistry.UnpackAny(v, &acc)
		if v != nil {
			accounts = append(accounts, acc)
		}
	}
	return accounts, nil
}

func GetAccountCoins(
	address string,
) (sdk.Coins, error) {
	// PrintPayload("GetAccountCoins", address)
	resp, err := bankClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: address,
	})
	logrus.Debug(resp.String())
	if err != nil {
		logrus.Error("Can't get account coin")
		return nil, err
	}
	return resp.Balances, nil
}
func PrintBaseAccountInfo(addrs ...string) {
	for _, addr := range addrs {
		account, err := GetAccountInfo(addr)
		if err != nil {
			logrus.Error("Account not found: ", err)
		} else {
			accountCoin, err := GetAccountCoins(addr)
			if err != nil {
				logrus.Error("Account coin not found: ", err)
			} else {
				logrus.Info("Account Info Of Address: ", addr)
				logrus.Info("Account Number: ", account.GetAccountNumber())
				logrus.Info("Account Sequence: ", account.GetSequence())
				logrus.Info("Account Denoms: ", accountCoin.Denoms())
				logrus.Info("Account Coin: ", accountCoin.String())
			}
		}
	}
}

//export QueryAccount
func QueryAccount(address string) (*BaseAccount, error) {
	// PrintPayload("QueryAccount", address)
	account, err := GetAccountInfo(address)
	if err != nil {
		logrus.Error("Account not found: ", err)
		return nil, err
	}
	accountCoin, err := GetAccountCoins(address)
	if err != nil {
		logrus.Error("Account coin not found: ", err)
		return nil, err
	}

	addressBytes, _ := sdk.AccAddressFromBech32(address)
	baseAccount := BaseAccount{
		Address:       addressBytes.Bytes(),
		Coins:         &Coins{Length: uint64(len(accountCoin)), Array: make([]*Coin, len(accountCoin))},
		PubKey:        account.GetPubKey().Bytes(),
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
	}

	for i, coin := range accountCoin {
		baseAccount.Coins.Array[i] = &Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.Uint64(),
		}
	}
	return &baseAccount, nil
}

//export NewNibiruClientDefault
func NewNibiruClientDefault() int {
	logrus.Println("Call [NewNibiruClientDefault]") // Use logrus instead of fmt.Println
	grpcConn, err := gonibi.GetGRPCConnection(networkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		logrus.Println("[NewNibiruClientDefault] GetGRPCConnection error: " + err.Error())
		return Fail
	}

	gosdk, err := gonibi.NewNibiruClient(networkInfo.ChainID, grpcConn, networkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.Println("[NewNibiruClientDefault] Connect to network error: " + err.Error())
		return Fail
	}
	logrus.Println("[NewNibiruClientDefault] Connected to " + gosdk.ChainId)
	return Success
}

//export NewNibiruClient
func NewNibiruClient(chainId string, grpcEndpoint string, rpcEndpoint string) int {
	logrus.Println("Call [NewNibiruClient]")
	grpcConn, err := gonibi.GetGRPCConnection(grpcEndpoint, true, 2)
	if err != nil {
		logrus.Println("[NewNibiruClient] GetGRPCConnection error: " + err.Error())
		return Fail
	}

	gosdk, err := gonibi.NewNibiruClient(chainId, grpcConn, rpcEndpoint)
	if err != nil {
		logrus.Println("[NewNibiruClient] Connect to network error: " + err.Error())
		return Fail
	}

	logrus.Println("[NewNibiruClient] Connected to " + gosdk.ChainId)
	return Success
}

//export GenerateRecoveryPhrase
func GenerateRecoveryPhrase() string {
	logrus.Info("Call GenerateRecoveryPhrase")
	const mnemonicEntropySize = 256
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		logrus.Error("Can't generate recovery phrase")
		return ""
	}
	phrase, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		logrus.Error("Can't generate recovery phrase")
		return ""
	}
	logrus.Info("Return recovery phrase: ", phrase)
	return phrase
}

//export CreateAccount
func CreateAccount(keyName string, mnemonic string, passphase string) int {
	// PrintPayload("CreateAccount", keyName, mnemonic, passphase)
	record, _, err := gonibi.CreateSigner(mnemonic, gosdk.Keyring, keyName)
	if err != nil {
		logrus.Debug("Failed to create new account", err)
		return Fail
	}

	addr, _ := record.GetAddress()
	logrus.Printf("Account created with address: %s", addr.String())

	PrintListSigners()
	return Success
}

//export GetPrivKeyFromMnemonic
func GetPrivKeyFromMnemonic(mnemoic string, keyName string) []byte {
	// PrintPayload("GetPrivKeyFromMnemonic", mnemoic, keyName)
	kring := gosdk.Keyring
	privKey, _, err := gonibi.PrivKeyFromMnemonic(kring, mnemoic, keyName)
	if err != nil {
		logrus.Debug("Failed to get priv key", err)
		return []byte{}
	}
	logrus.Info("Address String", privKey.PubKey().Address().String())
	return privKey.Bytes()
}

//export GetAddressFromKeyName
func GetAddressFromKeyName(keyName string) string {
	// PrintPayload("GetAddressFromKeyName", keyName)
	keyInfo, err := gosdk.Keyring.Key(keyName)
	if err != nil {
		logrus.Error("Failed to get address", err)
		return ""
	}
	addr, err := keyInfo.GetAddress()
	if err != nil {
		logrus.Error("Failed to get address", err)
		return ""
	}

	logrus.Info("Return Address: ", addr.String())

	return addr.String()
}

//export ImportAccountFromMnemoic
func ImportAccountFromMnemoic(mnemonic string, keyName string) int {
	mnemonicStr := mnemonic
	keyNameStr := keyName
	// PrintPayload("ImportAccountFromMnemoic", mnemonicStr, keyNameStr)
	// Create a keyring
	kring := gosdk.Keyring
	signer, _, err := gonibi.CreateSigner(mnemonicStr, kring, keyNameStr)
	if err != nil {
		logrus.Debug("Failed to import account:", err)
		return Fail
	}
	logrus.Printf("Susscess to import account: name: %s", signer.Name)
	return Success
}

//export ImportAccountFromPrivateKey
func ImportAccountFromPrivateKey(privateKey []byte, keyName string) int {
	// PrintPayload("ImportAccountFromPrivateKey", keyName)
	// Create a PrivKey instance and assign the decoded bytes to its Key field
	privKey := secp256k1.PrivKey{
		Key: privateKey,
	}
	// Create a keyring
	signer, err := gonibi.CreateSignerFromPrivKey(&privKey, keyName)
	if err != nil {
		return Fail
	}
	logrus.Info("Success to import account: ", signer.Name)
	return Success
}

//export GetListAccount
// func GetListAccount(length *int) KeyInfo {
// 	logrus.Debug("Call GetListAccount")
// 	signers, err := gosdk.Keyring.List()
// 	if err != nil {
// 		logrus.Debug("Error can't get list signer:", err)
// 		return KeyInfo{}
// 	}

//		return signers
//	}
//

// ConvertKeyInfo converts a keyring.Record to a KeyInfo struct
func convertKeyInfo(key *keyring.Record) *KeyInfo {
	// Create a new KeyInfo struct
	keyInfo := &KeyInfo{}

	// Set fields in the KeyInfo struct
	keyInfo.Type = uint32(key.GetType())
	keyInfo.Name = key.Name

	// Copy the public key bytes
	pubkey, _ := key.GetPubKey()
	pubKeyBytes := pubkey.Bytes()
	copy(keyInfo.PubKey[:], pubKeyBytes)

	// Copy the address bytes
	address, err := key.GetAddress()
	if err != nil {
		logrus.Error("Can't get address")
		return nil
	}
	addressBytes := address.Bytes()
	copy(keyInfo.Address[:], addressBytes)

	// Return the KeyInfo struct
	return keyInfo
}

//export GetAccountByKeyName
func GetAccountByKeyName(keyName string) *KeyInfo {
	// PrintPayload("GetAccountByKeyName", keyName)
	keyInfo, err := gosdk.Keyring.Key(keyName)
	if err != nil {
		logrus.Error("GetAccountByKeyName Failed: ", err)
		return nil
	}

	logrus.Debug("Account find: ")
	addr, err := keyInfo.GetAddress()
	if err != nil {
		logrus.Error("GetAccountByKeyName Failed to get dddress: ", err)
		return nil
	}
	logrus.Infof("Name: %s\nPubkey: %s\n address: %s", keyInfo.Name, keyInfo.PubKey, addr.String())
	return convertKeyInfo(keyInfo)
}

//export GetAccountByAddress
func GetAccountByAddress(addr string) *KeyInfo {
	// PrintPayload("GetAccountByAddress", addr)
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		logrus.Error("GetAccountByaddr Failed: ", err)
		return nil
	}
	logrus.Printf("C address: %s, niburu address: %s", addr, address)
	keyInfo, err := gosdk.Keyring.KeyByAddress(address)
	if err != nil {
		logrus.Error("GetAccountByaddr Failed: ", err)
		return nil
	}

	return convertKeyInfo(keyInfo)
}

//export HasKeyByName
func HasKeyByName(name string) int {
	logrus.Debug("Call HasKeyByName")
	has, err := gosdk.Keyring.Key(name)
	if err != nil {
		logrus.Error("HasKeyByName Fail: ", err)
		return Fail
	}

	if has != nil {
		return Success
	} else {
		return Fail
	}
}

//export HasKeyByAddress
func HasKeyByAddress(addr string) int {
	logrus.Debug("Call HasKeyByAddres")
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return Fail
	}

	a, err := gosdk.Keyring.KeyByAddress(address)
	if err != nil {
		logrus.Error("GetAccountByAddr Fail: ", err)
		return Fail
	}

	if a != nil {
		logrus.Debug("Key Name: ", a.Name)
		return Success
	} else {
		return Fail
	}
}

func PrintListSigners() {
	logrus.Debug("Call GetListAccount")
	signers, err := gosdk.Keyring.List()
	if err != nil {
		logrus.Debug("Error can't get list signer:", err)
	}

	for _, signer := range signers {

		addr, err := signer.GetAddress()
		if err != nil {
			logrus.Error("GetAccountByKeyName Failed to get address: ", err)
		}
		logrus.Infof("Name: %s\n address: %s", signer.Name, addr.String())
	}
}

//export DeleteAccount
func DeleteAccount(keyName string, password string) int {
	logrus.Debug("Call DeleteAccount")

	err := gosdk.Keyring.Delete(keyName)
	if err != nil {
		logrus.Debug("Error:", err)
		return Fail
	}
	PrintListSigners()
	return Success
}

//export TransferToken
func TransferToken(fromAddress, toAddress, denom string, amount int) int {
	logrus.Info("Call TransferToken")
	// PrintPayload("TransferToken", fromAddress, toAddress, denom, amount)
	// Convert C strings to Go strings
	fromStr := fromAddress
	toStr := toAddress
	denomStr := denom
	// check if from and to account is nil
	fromAcc, err := QueryAccount(fromStr)
	if err != nil || fromAcc == nil {
		logrus.Error("Can't get fromAccount", err)
		return Fail
	}
	toAcc, err := QueryAccount(toStr)
	if toAcc == nil || err != nil {
		logrus.Error("Can't get toAccount", err)
		return Fail
	}

	// Get the sender's address
	from, err := sdk.AccAddressFromBech32(fromStr)
	if err != nil {
		logrus.Error("Can't get fromAddress", err)
		return Fail
	}

	// Get the recipient's address
	to, err := sdk.AccAddressFromBech32(toStr)
	if err != nil {
		logrus.Error("Can't get toAddress", err)
		return Fail
	}

	// Create a coin with the specified denomination and amount
	coin := sdk.NewCoins(sdk.NewInt64Coin(denomStr, int64(amount)))

	// Create a MsgSend message to transfer tokens
	msgSend := banktypes.NewMsgSend(from, to, coin)
	defer PrintBaseAccountInfo(fromStr, toStr)
	// Broadcast the transaction to the blockchain network
	_, err = gosdk.BroadcastMsgs(from, msgSend)

	if err != nil {
		logrus.Error("Error BroadcastMsgs", err)
		return Fail
	}

	return Success
}

//export ExecuteWasmContract
func ExecuteWasmContract(senderAddress, contractAddress, executeMsg, denom string, amount int) string {
	// Convert C types to Go types
	fromStr := senderAddress
	contractStr := contractAddress
	msgStr := executeMsg
	denomStr := denom
	amountInt := sdk.NewInt(int64(amount))

	// Get the sender's address
	from, err := sdk.AccAddressFromBech32(fromStr)
	if err != nil {
		logrus.Error("Failed to parse sender address:", err)
		return ""
	}

	// Get the contract address
	contract, err := sdk.AccAddressFromBech32(contractStr)
	if err != nil {
		logrus.Error("Failed to parse contract address:", err)
		return ""
	}

	// Create the coins to send with the message
	coins := sdk.NewCoins(sdk.NewCoin(denomStr, amountInt))

	// Create the Wasm execute message
	msgExe := &wasmtypes.MsgExecuteContract{
		Sender:   from.String(),
		Contract: contract.String(),
		Msg:      []byte(msgStr),
		Funds:    coins,
	}

	// Broadcast the transaction to the blockchain network
	responseMsg, err := gosdk.BroadcastMsgs(from, msgExe)

	if err != nil {
		logrus.Error("Error BroadcastMsgs", err)
		return "nil"
	}

	logrus.Info("Response: ", string(responseMsg.String()))

	return responseMsg.TxHash
}

func QueryWasmContract(contractAddress, queryMsg string) string {
	// PrintPayload("QueryWasmContract", contractAddress, queryMsg)
	// Convert C types to Go types
	contractStr := contractAddress
	msgStr := queryMsg

	// Get the contract address
	contract, err := sdk.AccAddressFromBech32(contractStr)
	if err != nil {
		logrus.Error("Failed to parse contract address:", err)
		return ""
	}

	// Create the Wasm execute message
	msgExe := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contract.String(),
		QueryData: []byte(msgStr),
	}

	responseMsg, err := wasmClient.SmartContractState(context.Background(), msgExe)

	if err != nil {
		logrus.Error("Error SmartContractState", err)
		return ""
	}

	logrus.Info("Response: ", responseMsg.String())

	return responseMsg.String()
}

//export QueryTXHash
func QueryTXHash(txHash string) string {
	// PrintPayload("QueryTXHash", txHash)
	decodedBytes, err := hex.DecodeString(txHash)

	if err != nil {
		logrus.Error("Error getTX info: ", err)
		return ""
	}

	resultTx, err := gosdk.CometRPC.Tx(context.Background(), decodedBytes, true)

	if err != nil {
		logrus.Error("Error getTX info: ", err)
		return ""
	}

	logrus.Info("Result: ", resultTx.TxResult.Log)
	return resultTx.TxResult.String()
}
