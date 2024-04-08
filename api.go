package main

/*
#include <stdint.h> // for uint32_t

// If crypto.Address and crypto.PubKey are fixed-size byte arrays, define their sizes
#define ADDRESS_SIZE 20 // Example size, adjust according to actual crypto.Address size
#define PUBKEY_SIZE  58 // Example size, adjust according to actual crypto.PubKey size

// Define a C-compatible KeyInfo struct
typedef struct {
	uint32_t Type;
	const char* Name;
	const uint8_t PubKey[PUBKEY_SIZE];
	const uint8_t Address[ADDRESS_SIZE];
} KeyInfo;

typedef struct {
	KeyInfo* Info;
	char* Password;
} UserAccount;

// Define the Coin type in C, assuming both Denom and Amount are strings
typedef struct {
    char *Denom;
    uint64_t Amount;
} Coin;

// If Coins is a dynamic array or slice of Coin, you will need a struct to represent it
typedef struct {
    Coin *Array;     // Pointer to the first Coin element
    size_t Length;   // Number of elements in the Coins array
} Coins;

// Then define the BaseAccount struct in C
typedef struct {
    uint8_t Address[ADDRESS_SIZE];
    Coins*   Coins;              // Assuming Coins is represented as above
    uint8_t PubKey[PUBKEY_SIZE];
    uint64_t AccountNumber;
    uint64_t Sequence;
} BaseAccount;
*/
import "C"
import (
	"context"
	"errors"
	"unsafe"

	"github.com/Unique-Divine/gonibi"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/sirupsen/logrus"
)

func main() {

}

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

type UserAccount struct {
	KeyInfo  keyring.Record
	Password string
}

var gosdk gonibi.NibiruClient
var sdkCtx sdk.Context
var authClient authtypes.QueryClient
var bankClient banktypes.QueryClient
var networkInfo NetworkInfo

func InitClients() error {
	authClient = authtypes.NewQueryClient(gosdk.Querier.ClientConn)
	bankClient = banktypes.NewQueryClient(gosdk.Querier.ClientConn)
	if authClient == nil || bankClient == nil {
		return errors.New("can't init client")
	}
	return nil
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
	logrus.Debug("Call GetAccountCoins")
	resp, err := bankClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: address,
	})
	logrus.Debug(resp.String())
	if err != nil {
		logrus.Error("Can't get account coin")
		return nil, err
	}
	logrus.Debug("End Call GetAccountCoins")
	return resp.Balances, nil
}

//export QueryAccount
func QueryAccount(address *C.char) *C.BaseAccount {
	logrus.Debug("Call QueryAccount")

	if _, err := GetListAccountInfo(); err != nil {
		logrus.Info(err)
	}

	addr, err := sdk.AccAddressFromBech32(C.GoString(address))
	if err != nil {
		logrus.Error("GetAccountByaddr Failed: ", err)
		return nil
	}
	logrus.Info("QueryAccount ~ addr:", addr.String())
	account, err := GetAccountInfo(C.GoString(address))
	if err != nil {
		logrus.Error("Account not found: ", err)
		return nil
	}

	logrus.Info("QueryAccount ~ Account:", account)
	// Allocate memory for BaseAccount in C.
	cAccount := (*C.BaseAccount)(C.malloc(C.sizeof_BaseAccount))
	if cAccount == nil {
		// Handle allocation failure if needed
		return nil
	}

	// Allocate memory for Coins in C.
	cAccount.Coins = (*C.Coins)(C.malloc(C.sizeof_Coins))
	if cAccount.Coins == nil {
		// Handle allocation failure if needed
		// C.free(unsafe.Pointer(cAccount))
		return nil
	}
	// get account coin

	accountCoins, err := GetAccountCoins(addr.String())
	if err != nil {
		logrus.Error("Can't get account coins")
		return nil
	}
	cAccount.Coins.Length = C.size_t(len(accountCoins))
	cAccount.Coins.Array = (*C.Coin)(C.malloc(C.sizeof_Coin * cAccount.Coins.Length))
	if cAccount.Coins.Array == nil {
		// Handle allocation failure if needed
		// C.free(unsafe.Pointer(cAccount.Coins))
		// C.free(unsafe.Pointer(cAccount))
		return nil
	}

	cCoinPtr := cAccount.Coins.Array
	for _, coin := range accountCoins {
		// Allocate and set the C string equivalents
		cCoinPtr.Denom = C.CString(coin.Denom)
		cCoinPtr.Amount = C.uint64_t(coin.Amount.Int64())
		// Move the pointer to the next array element; this is equivalent to incrementing an array index
		cCoinPtr = (*C.Coin)(unsafe.Pointer(uintptr(unsafe.Pointer(cCoinPtr)) + C.sizeof_Coin))
	}

	// Copy the account address bytes to the C struct.
	addressBytes := account.GetAddress().Bytes()
	if len(addressBytes) > len(cAccount.Address) {
		// Handle error: the address is too big for the allocated array.
		// Remember to free all previously allocated memory.
		// C.free(unsafe.Pointer(cAccount.Coins.Array))
		// C.free(unsafe.Pointer(cAccount.Coins))
		// C.free(unsafe.Pointer(cAccount))
		return nil
	}
	for i, b := range addressBytes {
		cAccount.Address[i] = C.uint8_t(b)
	}

	// Copy the public key bytes to the C struct if a public key is present.
	if account.GetPubKey() != nil {
		pubKeyBytes := account.GetPubKey().Bytes()
		if len(pubKeyBytes) > len(cAccount.PubKey) {
			// Handle error: the public key is too big for the allocated array.
			// Remember to free all previously allocated memory.
			// C.free(unsafe.Pointer(cAccount.Coins.Array))
			// C.free(unsafe.Pointer(cAccount.Coins))
			// C.free(unsafe.Pointer(cAccount))
			return nil
		}
		for i, b := range pubKeyBytes {
			cAccount.PubKey[i] = C.uint8_t(b)
		}
	}

	cAccount.AccountNumber = C.uint64_t(account.GetAccountNumber())
	cAccount.Sequence = C.uint64_t(account.GetSequence())

	return cAccount
}

//export NewNibiruClientDefault
func NewNibiruClientDefault() C.int {
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
func NewNibiruClient(chainId *C.char, grpcEndpoint *C.char, rpcEndpoint *C.char) C.int {
	logrus.Println("Call [NewNibiruClient]")
	grpcConn, err := gonibi.GetGRPCConnection(C.GoString(grpcEndpoint), true, 2)
	if err != nil {
		logrus.Println("[NewNibiruClient] GetGRPCConnection error: " + err.Error())
		return Fail
	}

	gosdk, err := gonibi.NewNibiruClient(C.GoString(chainId), grpcConn, C.GoString(rpcEndpoint))
	if err != nil {
		logrus.Println("[NewNibiruClient] Connect to network error: " + err.Error())
		return Fail
	}

	logrus.Println("[NewNibiruClient] Connected to " + gosdk.ChainId)
	return Success
}

//export GenerateRecoveryPhrase
func GenerateRecoveryPhrase() *C.char {
	const mnemonicEntropySize = 256
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		return C.CString("")
	}
	phrase, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		return C.CString("")
	}
	return C.CString(phrase)
}

// ToCKeyInfo converts KeyInfo to its C representation.
// todo: currently it's wrong
func convertKeyInfo(key *keyring.Record) *C.KeyInfo {
	// Allocate memory for KeyInfo in C.
	cKeyInfo := (*C.KeyInfo)(C.malloc(C.sizeof_KeyInfo))
	if cKeyInfo == nil {
		// Handle allocation failure if needed
		return nil
	}

	// Set fields in the KeyInfo struct.
	cKeyInfo.Type = C.uint32_t(key.GetType())
	cKeyInfo.Name = C.CString(key.Name) // This will need to be freed in C.

	// Copy the public key bytes.
	pubkey, err := key.GetPubKey()
	if err != nil {
		logrus.Error("Can't get public key")
		return nil
	}

	pubKeyBytes := pubkey.Bytes()

	if len(pubKeyBytes) > len(cKeyInfo.PubKey) {
		// Handle error: the address is too big for the allocated array.
		// C.free(unsafe.Pointer(cKeyInfo.Name))
		// C.free(unsafe.Pointer(cKeyInfo))
		return nil
	}
	for i, b := range pubKeyBytes {
		cKeyInfo.PubKey[i] = C.uint8_t(b)
	}

	// Copy the address bytes.
	address, err := key.GetAddress()
	if err != nil {
		logrus.Error("Can't get public key", err)
		return nil
	}
	addressBytes := address.Bytes()
	if len(addressBytes) > len(cKeyInfo.Address) {
		// Handle error: the address is too big for the allocated array.
		// C.free(unsafe.Pointer(cKeyInfo.Name))
		// C.free(unsafe.Pointer(cKeyInfo))
		return nil
	}
	for i, b := range addressBytes {
		cKeyInfo.Address[i] = C.uint8_t(b)
	}

	// Return the heap-allocated KeyInfo.
	return cKeyInfo
}

//export CreateAccount
func CreateAccount(keyName *C.char, mnemonic *C.char, passphase *C.char) C.int {
	logrus.Debug("Call Creating Account keyname:", C.GoString(keyName), "mnemonic:", C.GoString(mnemonic))
	algo := hd.Secp256k1
	// Create a keyring
	record, err := gosdk.Keyring.NewAccount(C.GoString(keyName), C.GoString(mnemonic), C.GoString(passphase), sdk.GetConfig().GetFullBIP44Path(), algo)
	if err != nil {
		logrus.Debug("Failed to create new account", err)
		return Fail
	}
	logrus.Printf("Account created: %s, %s\n", record.Name, record.PubKey.String())
	return Success
}

func NewBaseAccount(keyInfo keyring.Record) (account *authtypes.BaseAccount, err error) {
	logrus.Info("Call NewBaseAccount")
	keyAddr, err := keyInfo.GetAddress()
	logrus.Info("Account Address: ", keyAddr)
	if err != nil {
		logrus.Error("Can't get address: ", err)
		return account, err
	}
	pubKey, err := keyInfo.GetPubKey()
	if err != nil {
		logrus.Error("Can't get pubkey: ", err)
		return account, err
	}
	var accNumber uint64 = 0
	var accSequence uint64 = 0
	acc := authtypes.NewBaseAccount(keyAddr, pubKey, accNumber, accSequence)

	return acc, nil
}

//export GetPrivKeyFromMnemonic
func GetPrivKeyFromMnemonic(mnemoic *C.char, keyName *C.char) *C.uint8_t {
	logrus.Debug("Call GetPrivKeyFromMnemonic")
	kring := gosdk.Keyring
	privKey, _, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), C.GoString(keyName))
	if err != nil {
		logrus.Debug("Failed to get priv key", err)
		return nil
	}
	logrus.Info("Address String", privKey.PubKey().Address())
	return revertToCData(privKey.Bytes())
}

// Revert C data and length to Go byte slice
func revertToCData(byteSlice []byte) *C.uint8_t {
	// Ensure the byte slice is not nil
	if byteSlice == nil {
		return nil
	}

	// Create a new C byte array
	cData := C.malloc(C.size_t(len(byteSlice)))
	if cData == nil {
		return nil
	}

	// Copy data from byte slice to C array
	cSlice := (*[1 << 30]byte)(cData)[:len(byteSlice):len(byteSlice)]
	copy(cSlice, byteSlice)
	// Defer the free operation to release the allocated memory

	return (*C.uint8_t)(cData)
}

// Convert a *C.uint8_t pointer to a Go byte slice
func cUint8ToGoSlice(cData *C.uint8_t) []byte {
	if cData == nil {
		return nil
	}

	// Calculate the length of the C data dynamically
	var length int
	for length = 0; *(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cData)) + uintptr(length))) != 0; length++ {
	}

	// Convert C data to Go slice
	goSlice := make([]byte, length)
	for i := 0; i < length; i++ {
		goSlice[i] = byte(*((*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cData)) + uintptr(i)))))
	}
	return goSlice
}

//export GetAddressFromKeyName
func GetAddressFromKeyName(keyName *C.char) *C.char {
	logrus.Println("Call GetAddressFromKeyName")
	keyInfo, err := gosdk.Keyring.Key(C.GoString(keyName))
	if err != nil {
		logrus.Debug("Failed to get address", err)
		return nil
	}
	addr, err := keyInfo.GetAddress()
	if err != nil {
		logrus.Debug("Failed to get address", err)
		return nil
	}

	logrus.Info("Return Address: ", addr.String())

	return C.CString(addr.String())
}

//export ImportAccountFromMnemoic
func ImportAccountFromMnemoic(mnemonic *C.char, keyName *C.char) C.int {
	logrus.Debug("Call Import Account")
	// GetListAccountInfo()
	mnemonicStr := C.GoString(mnemonic)
	// Create a keyring
	kring := gosdk.Keyring
	signer, privateKey, err := gonibi.CreateSigner(mnemonicStr, kring, C.GoString(keyName))
	if err != nil {
		logrus.Debug("Failed to import account:", err)
		return Fail
	}
	if err := gonibi.AddSignerToKeyring(kring, privateKey, privateKey.PubKey().String()); err != nil {
		logrus.Error("Can't assing singer to keyring: ", err)
		return Fail
	}
	logrus.Printf("Susscess to import account: name: %s", signer.Name)
	return Success
}

//export ImportAccountFromPrivateKey
func ImportAccountFromPrivateKey(privateKey *C.uint8_t, keyName *C.char) C.int {
	logrus.Debug("Import Account")
	// Decode the private key string from hex
	privKeyBytes := cUint8ToGoSlice(privateKey)
	if privKeyBytes == nil {
		logrus.Debug("Can not get private key")
	}

	// Create a PrivKey instance and assign the decoded bytes to its Key field
	privKey := secp256k1.PrivKey{
		Key: privKeyBytes,
	}

	logrus.Info("Pubkey String: ", privKey.PubKey().String())
	// Create a keyring
	signer, err := gonibi.CreateSignerFromPrivKey(&privKey, C.GoString(keyName))
	logrus.Info("Import New Account Success", signer.String())
	if err != nil {
		return Fail
	}
	return Success
}

//export GetListAccount
func GetListAccount(length *C.int) **C.KeyInfo {
	logrus.Debug("Call GetListAccount")
	keys, err := gosdk.Keyring.List()
	if err != nil {
		*length = 0
		return nil
	}

	*length = C.int(len(keys))

	// Allocate memory for the array of pointers
	keyInfos := C.malloc(C.size_t(len(keys)) * C.size_t(unsafe.Sizeof(uintptr(0))))

	// Convert the allocated memory to **C.KeyInfo
	cKeyInfos := (**C.KeyInfo)(keyInfos)

	// Iterate over the keys and assign each pointer
	for i, key := range keys {
		// Allocate memory for the KeyInfo struct
		keyInfo := (*C.KeyInfo)(C.malloc(C.size_t(unsafe.Sizeof(C.KeyInfo{}))))

		// Assign values to the KeyInfo struct
		keyInfo.Name = C.CString(key.Name)
		// Assign other fields of KeyInfo struct as needed

		// Assign the pointer to the array
		(*[1 << 30]unsafe.Pointer)(unsafe.Pointer(keyInfos))[i] = unsafe.Pointer(keyInfo)
	}

	return cKeyInfos
}

//export GetAccountByKeyName
func GetAccountByKeyName(keyName *C.char) *C.KeyInfo {
	logrus.Debug("Call GetAccountByKeyName")
	keyInfo, err := gosdk.Keyring.Key(C.GoString(keyName))
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
	logrus.Printf("Name: %s\nPubkey: %s\n address: %s", keyInfo.Name, keyInfo.PubKey, addr.String())
	return convertKeyInfo(keyInfo)
}

//export GetAccountByAddress
func GetAccountByAddress(addr *C.char) *C.KeyInfo {
	logrus.Debug("Call GetAccountByAddress")
	address, err := sdk.AccAddressFromBech32(C.GoString(addr))
	if err != nil {
		logrus.Error("GetAccountByaddr Failed: ", err)
		return nil
	}
	logrus.Printf("C address: %s, niburu address: %s", C.GoString(addr), address)
	keyInfo, err := gosdk.Keyring.KeyByAddress(address)
	if err != nil {
		logrus.Error("GetAccountByaddr Failed: ", err)
		return nil
	}

	return convertKeyInfo(keyInfo)
}

//export HasKeyByName
func HasKeyByName(name *C.char) C.int {
	logrus.Debug("HasKeyByName called")
	has, err := gosdk.Keyring.Key(C.GoString(name))
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
func HasKeyByAddress(addr *C.char, len C.int) C.int {
	logrus.Debug("HasKeyByAddress called")
	address, err := sdk.AccAddressFromBech32(C.GoString(addr))
	if err != nil {
		return Fail
	}

	a, err := gosdk.Keyring.KeyByAddress(address)
	if err != nil {
		logrus.Error("GetAccountByAddr Fail: ", err)
		return Fail
	}

	if a != nil {
		logrus.Debug("keyInfor ", a.String())
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
		logrus.Printf("Name: %s\nPubkey: %s\n address: %s", signer.Name, signer.PubKey, addr)
	}
}

//export DeleteAccount
func DeleteAccount(keyName *C.char, password *C.char) C.int {
	logrus.Debug("Call DeleteAccount")

	err := gosdk.Keyring.Delete(C.GoString(keyName))
	if err != nil {
		logrus.Debug("Error:", err)
		return Fail
	}
	PrintListSigners()
	return Success
}

//export TestTransferToken
func TestTransferToken() C.int {
	// accounts, err := GetListAccountInfo()
	// if err != nil {
	// 	return Fail
	// }
	addr1 := "nibi1zaavvzxez0elundtn32qnk9lkm8kmcsz44g7xl"
	acc1Coin, err := GetAccountCoins(addr1)
	if err != nil {
		return Fail
	}
	logrus.Info(addr1, " ", acc1Coin.Denoms())
	addr2 := "nibi1tq2ynj3x3k7z09r2z6qqndvpmk57ffr2ma39nh"
	acc2Coin, err := GetAccountCoins(addr2)
	if err != nil {
		return Fail
	}
	logrus.Info(addr2, " ", acc2Coin.Denoms(), acc2Coin.AmountOf("unibi"))
	denomStr := "unibi"

	coin := sdk.NewCoins(sdk.NewInt64Coin(denomStr, 100))

	// Create a MsgSend message to transfer tokens
	rawaddr1, _ := sdk.AccAddressFromBech32(addr1)
	rawaddr2, _ := sdk.AccAddressFromBech32(addr2)
	msgSend := banktypes.NewMsgSend(rawaddr1, rawaddr2, coin)

	// Broadcast the transaction to the blockchain network
	txRsp, err := gosdk.BroadcastMsgs(rawaddr1, msgSend)
	if err != nil || txRsp == nil {
		logrus.Error("Transfer Error", err)
		return Fail
	}

	PrintAccount()
	return Success
}
func PrintAccount() error {
	accounts, err := GetListAccountInfo()
	if err != nil {
		return err
	}
	addr1 := accounts[1].GetAddress()
	acc1Coin, err := GetAccountCoins(addr1.String())
	if err != nil {
		return err
	}
	logrus.Info(addr1.String(), " ", acc1Coin.String())
	addr2 := accounts[2].GetAddress()
	acc2Coin, err := GetAccountCoins(addr2.String())
	if err != nil {
		return err
	}
	logrus.Info(addr2.String(), " ", acc2Coin.Denoms(), acc2Coin.AmountOf("unibi"))
	return nil
}

func TransferToken(fromAddress, toAddress, denom *C.char, amount C.int) (*sdk.TxResponse, error) {
	// Convert C strings to Go strings
	fromStr := C.GoString(fromAddress)
	toStr := C.GoString(toAddress)
	denomStr := C.GoString(denom)

	// Get the sender's address
	from, err := sdk.AccAddressFromBech32(fromStr)
	if err != nil {
		logrus.Error("Can't get fromAddress", err)
		return nil, err
	}

	// Get the recipient's address
	to, err := sdk.AccAddressFromBech32(toStr)
	if err != nil {
		logrus.Error("Can't get toAddress", err)
		return nil, err
	}

	// Create a coin with the specified denomination and amount
	coin := sdk.NewCoins(sdk.NewInt64Coin(denomStr, int64(amount)))

	// Create a MsgSend message to transfer tokens
	msgSend := banktypes.NewMsgSend(from, to, coin)

	// Broadcast the transaction to the blockchain network
	return gosdk.BroadcastMsgs(from, msgSend)
}
