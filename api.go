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
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"unsafe"

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

func main() {

}

// Constants to represent the success or failure of functions.
const (
	Success = 0 // Success indicates that the function completed its task without errors.
	Fail    = 1 // Fail indicates that the function encountered an error and did not complete successfully.
)

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

type UserAccount struct {
	KeyInfo  keyring.Record
	Password string
}

var gosdk gonibi.NibiruClient
var sdkCtx sdk.Context
var authClient authtypes.QueryClient
var bankClient banktypes.QueryClient
var wasmClient wasmtypes.QueryClient
var networkInfo NetworkInfo

func InitClients() error {
	// Initialize clients for respective services using the global gosdk instance
	authClient = authtypes.NewQueryClient(gosdk.Querier.ClientConn)
	bankClient = banktypes.NewQueryClient(gosdk.Querier.ClientConn)
	wasmClient = wasmtypes.NewQueryClient(gosdk.Querier.ClientConn)

	// Check if any client initialization failed
	if authClient == nil || bankClient == nil || wasmClient == nil {
		logrus.Error("Failed to initialize one or more clients")
		return errors.New("can't init client")
	}
	logrus.Info("Clients initialized successfully")
	return nil
}

func PrintPayload(funcName string, args ...interface{}) {
	logrus.WithField("Function", funcName).Info("Function call started")

	// Log each argument passed to the function
	for i, arg := range args {
		logrus.WithFields(logrus.Fields{
			"argument index": i,
			"value":          fmt.Sprintf("%v", arg),
		}).Debug("Function parameter")
	}
}

func init() {
	// Set up the logging format and level.
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	// Set the initial network information to local settings.
	networkInfo = LocalNetworkInfo

	// Attempt to establish a gRPC connection using the local network configuration.
	grpcConn, err := gonibi.GetGRPCConnection(networkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to initialize gRPC connection with endpoint %s", networkInfo.GrpcEndpoint)
	} else {
		logrus.WithField("endpoint", networkInfo.GrpcEndpoint).Info("gRPC connection established successfully")
	}

	// Initialize the Nibiru client with the obtained gRPC connection.
	gosdk, err = gonibi.NewNibiruClient(networkInfo.ChainID, grpcConn, networkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to initialize Nibiru client for chain ID %s", networkInfo.ChainID)
	} else {
		logrus.WithFields(logrus.Fields{
			"chainID":     networkInfo.ChainID,
			"rpcEndpoint": networkInfo.TmRpcEndpoint,
		}).Info("Nibiru client initialized successfully")
	}

	// Initialize clients for auth, bank, and wasm modules.
	if err := InitClients(); err != nil {
		logrus.WithError(err).Error("Failed to initialize clients")
	} else {
		logrus.Info("All clients initialized successfully")
	}

	// Confirm the package initialization is complete.
	logrus.Info("Package initialization completed successfully")
}

// SwitchNetwork changes the network configuration based on the provided network name.
// It returns Success (0) if the switch is successful, otherwise Fail (1).
//
//export SwitchNetwork
func SwitchNetwork(network *C.char) C.int {
	// Convert C string to Go string
	networkStr := C.GoString(network)
	logrus.WithField("network", networkStr).Info("Attempting to switch network")

	// Determine the appropriate network settings based on the input
	var grpcInsecure bool
	switch networkStr {
	case "local":
		networkInfo = LocalNetworkInfo
		grpcInsecure = true
	case "dev":
		networkInfo = DevNetworkInfo
		grpcInsecure = false
	case "test":
		networkInfo = TestNetworkInfo
		grpcInsecure = false
	case "main":
		networkInfo = MainNetworkInfo
		grpcInsecure = false
	default:
		logrus.WithField("network", networkStr).Warn("Unknown network specified, defaulting to test network")
		networkInfo = TestNetworkInfo
		grpcInsecure = false
	}

	// Establish a new gRPC connection with the updated network settings
	grpcConn, err := gonibi.GetGRPCConnection(networkInfo.GrpcEndpoint, grpcInsecure, 2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": networkInfo.GrpcEndpoint,
			"error":    err,
		}).Error("Failed to initialize gRPC connection")
		return Fail
	}

	// Initialize the Nibiru client with the new gRPC connection
	gosdk, err = gonibi.NewNibiruClient(networkInfo.ChainID, grpcConn, networkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"chainID":     networkInfo.ChainID,
			"rpcEndpoint": networkInfo.TmRpcEndpoint,
			"error":       err,
		}).Error("Failed to initialize Nibiru client")
		return Fail
	}

	logrus.WithField("network", networkStr).Info("Network switched successfully")

	// Reinitialize the clients to ensure they use the new network configuration
	if err := InitClients(); err != nil {
		logrus.WithError(err).Error("Failed to initialize clients after network switch")
		return Fail
	}

	return Success
}

// Niburu method
// GetAccountInfo retrieves the account information for a given blockchain address.
// It returns an AccountI interface and any error encountered during the retrieval or unpacking process.
func GetAccountInfo(address string) (authtypes.AccountI, error) {
	// Log the function call with the address parameter for debugging purposes.
	PrintPayload("GetAccountInfo", address)

	// Create a context with a timeout to ensure the request does not hang indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Request account information from the blockchain using the address.
	acc, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"address": address,
			"error":   err,
		}).Error("Failed to get account info from the blockchain")
		return nil, err
	}

	// Unpack the account information from the blockchain's response into a generic interface.
	var account authtypes.AccountI
	err = gosdk.EncCfg.InterfaceRegistry.UnpackAny(acc.Account, &account)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"address": address,
			"error":   err,
		}).Error("Failed to unpack account info")
		return nil, err
	}

	// Log the successfully fetched and unpacked account information.
	logrus.WithFields(logrus.Fields{
		"account number": account.GetAccountNumber(),
		"sequence":       account.GetSequence(),
		"address":        address,
	}).Info("Account information retrieved and unpacked successfully")

	return account, nil
}

// GetListAccountInfo retrieves a list of all accounts from the blockchain.
// It returns a slice of AccountI interfaces and any error encountered during the operation.
func GetListAccountInfo() ([]authtypes.AccountI, error) {
	// Initialize the query client using the existing gRPC connection.
	queryClient := authtypes.NewQueryClient(gosdk.Querier.ClientConn)

	// Create a context with a timeout to prevent the function from hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Request a list of all accounts from the blockchain.
	resp, err := queryClient.Accounts(ctx, &authtypes.QueryAccountsRequest{})
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch accounts from the blockchain")
		return nil, err
	}

	// Initialize a slice to hold the unpacked accounts.
	accounts := make([]authtypes.AccountI, 0, len(resp.Accounts))

	// Unpack each account and append to the accounts slice.
	for _, v := range resp.Accounts {
		var acc authtypes.AccountI
		if err := gosdk.EncCfg.InterfaceRegistry.UnpackAny(v, &acc); err != nil {
			logrus.WithFields(logrus.Fields{
				"account": v,
				"error":   err,
			}).Error("Failed to unpack account info")
			continue // Skip this account and continue with others
		}
		accounts = append(accounts, acc)
	}

	logrus.WithField("totalAccounts", len(accounts)).Info("All accounts fetched and unpacked successfully")
	return accounts, nil
}

// GetAccountCoins retrieves all coin balances associated with a given blockchain address.
// It returns a list of coins and any error encountered during the operation.
func GetAccountCoins(address string) (sdk.Coins, error) {
	// Log the function call with the address parameter to trace execution flow.
	PrintPayload("GetAccountCoins", address)

	// Create a context with a timeout to ensure the request does not hang indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Request the balance information from the blockchain using the address.
	resp, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: address,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"address": address,
			"error":   err,
		}).Error("Failed to retrieve account balances")
		return nil, err
	}

	// Log the response for debugging purposes. This is especially useful for verifying the correct response format.
	logrus.WithFields(logrus.Fields{
		"address":  address,
		"balances": resp.Balances.String(),
	}).Debug("Account balances retrieved successfully")

	return resp.Balances, nil
}

// PrintBaseAccountInfo fetches and logs basic information for a list of blockchain addresses.
func PrintBaseAccountInfo(addrs ...string) {
	// Iterate through all provided addresses
	for _, addr := range addrs {
		// Fetch account information using the address
		account, err := GetAccountInfo(addr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"address": addr,
				"error":   err,
			}).Error("Failed to get account information")
			continue // Skip to the next address if the current one fails
		}

		// Fetch coin balances associated with the account
		accountCoin, err := GetAccountCoins(addr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"address": addr,
				"error":   err,
			}).Error("Failed to get account coin balances")
			continue // Skip to the next address if the current one fails
		}

		// Log the fetched account information
		logrus.WithFields(logrus.Fields{
			"address":      addr,
			"accountNum":   account.GetAccountNumber(),
			"sequence":     account.GetSequence(),
			"coinDenoms":   accountCoin.Denoms(),
			"coinBalances": accountCoin.String(),
		}).Info("Account information retrieved successfully")
	}
}

// QueryAccount retrieves detailed account information given a blockchain address and returns a C-compatible structure.
//
//export QueryAccount
func QueryAccount(address *C.char) *C.BaseAccount {
	// Log the function call with the address parameter to trace execution flow.
	PrintPayload("QueryAccount", C.GoString(address))

	// Convert C string to Go string and validate the address format.
	addrStr := C.GoString(address)
	_, err := sdk.AccAddressFromBech32(addrStr)
	if err != nil {
		logrus.WithError(err).WithField("address", addrStr).Error("Invalid blockchain address format")
		return nil
	}

	// Fetch the account information from the blockchain.
	account, err := GetAccountInfo(addrStr)
	if err != nil {
		logrus.WithError(err).WithField("address", addrStr).Error("Failed to get account information")
		return nil
	}

	// Log retrieved account information for confirmation.
	logrus.WithFields(logrus.Fields{
		"address":       addrStr,
		"accountNumber": account.GetAccountNumber(),
		"sequence":      account.GetSequence(),
	}).Info("Account information retrieved successfully")

	// Allocate memory for the BaseAccount structure in C.
	cAccount, err := allocateBaseAccount()
	if err != nil {
		logrus.WithError(err).Error("Failed to allocate memory for BaseAccount")
		return nil
	}

	// Fetch and set account coins information.
	if err := setAccountCoins(cAccount, addrStr); err != nil {
		logrus.WithError(err).Error("Failed to set account coins")
		freeBaseAccount(cAccount) // Ensure all allocated memory is freed on error
		return nil
	}

	// Copy essential account details to the C structure.
	if err := copyAccountDetails(cAccount, account); err != nil {
		logrus.WithError(err).Error("Failed to copy account details to C structure")
		freeBaseAccount(cAccount) // Ensure all allocated memory is freed on error
		return nil
	}

	return cAccount
}

// Helper functions used within QueryAccount for specific tasks.
func allocateBaseAccount() (*C.BaseAccount, error) {
	cAccount := (*C.BaseAccount)(C.malloc(C.sizeof_BaseAccount))
	if cAccount == nil {
		return nil, errors.New("memory allocation failed for BaseAccount")
	}
	return cAccount, nil
}

func setAccountCoins(cAccount *C.BaseAccount, address string) error {
	accountCoins, err := GetAccountCoins(address)
	if err != nil {
		return err
	}

	cAccount.Coins = (*C.Coins)(C.malloc(C.sizeof_Coins))
	if cAccount.Coins == nil {
		return errors.New("memory allocation failed for Coins")
	}

	cAccount.Coins.Length = C.size_t(len(accountCoins))
	cAccount.Coins.Array = (*C.Coin)(C.malloc(C.sizeof_Coin * cAccount.Coins.Length))
	if cAccount.Coins.Array == nil {
		return errors.New("memory allocation failed for Coin array")
	}

	// Populate coin details in the allocated C array.
	populateCoins(cAccount.Coins.Array, accountCoins)
	return nil
}

func populateCoins(cCoinPtr *C.Coin, coins sdk.Coins) {
	for _, coin := range coins {
		cCoinPtr.Denom = C.CString(coin.Denom)
		cCoinPtr.Amount = C.uint64_t(coin.Amount.Int64())
		cCoinPtr = (*C.Coin)(unsafe.Pointer(uintptr(unsafe.Pointer(cCoinPtr)) + C.sizeof_Coin))
	}
}

func copyAccountDetails(cAccount *C.BaseAccount, account authtypes.AccountI) error {
	copyBytes(cAccount.Address[:], account.GetAddress().Bytes())
	if account.GetPubKey() != nil {
		copyBytes(cAccount.PubKey[:], account.GetPubKey().Bytes())
	}
	cAccount.AccountNumber = C.uint64_t(account.GetAccountNumber())
	cAccount.Sequence = C.uint64_t(account.GetSequence())
	return nil
}

func copyBytes(dest []C.uint8_t, src []byte) {
	for i, b := range src {
		dest[i] = C.uint8_t(b)
	}
}

func freeBaseAccount(cAccount *C.BaseAccount) {
	// Implement the logic to free all allocated memory associated with a BaseAccount.
}

// NewNibiruClientDefault initializes a new Nibiru client using default network settings.
// It returns Success if the client is successfully initialized, otherwise Fail.
//
//export NewNibiruClientDefault
func NewNibiruClientDefault() C.int {
	// Log the function call to trace execution.
	logrus.Info("Initializing Nibiru client with default settings")

	// Establish a gRPC connection using the default network information.
	grpcConn, err := gonibi.GetGRPCConnection(networkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"endpoint": networkInfo.GrpcEndpoint,
			"error":    err,
		}).Error("Failed to establish gRPC connection")
		return Fail
	}

	// Initialize the Nibiru client with the obtained gRPC connection.
	gosdk, err = gonibi.NewNibiruClient(networkInfo.ChainID, grpcConn, networkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"chainID":     networkInfo.ChainID,
			"rpcEndpoint": networkInfo.TmRpcEndpoint,
			"error":       err,
		}).Error("Failed to initialize Nibiru client")
		return Fail
	}

	// Log successful connection.
	logrus.WithFields(logrus.Fields{
		"chainID": networkInfo.ChainID,
	}).Info("Successfully connected to the Nibiru network")

	return Success
}

// NewNibiruClient initializes a new Nibiru client based on the provided chain ID, gRPC, and RPC endpoints.
// It returns Success if the client initialization is successful, otherwise Fail.
//
//export NewNibiruClient
func NewNibiruClient(chainId *C.char, grpcEndpoint *C.char, rpcEndpoint *C.char) C.int {
	// Convert C strings to Go strings for better handling in Go functions.
	chainIDStr := C.GoString(chainId)
	grpcEndpointStr := C.GoString(grpcEndpoint)
	rpcEndpointStr := C.GoString(rpcEndpoint)

	// Log the initiation of a new Nibiru client with specific network settings.
	logrus.WithFields(logrus.Fields{
		"chainID":      chainIDStr,
		"grpcEndpoint": grpcEndpointStr,
		"rpcEndpoint":  rpcEndpointStr,
	}).Info("Attempting to initialize new Nibiru client")

	// Establish a gRPC connection using the specified endpoint.
	grpcConn, err := gonibi.GetGRPCConnection(grpcEndpointStr, true, 2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"grpcEndpoint": grpcEndpointStr,
			"error":        err,
		}).Error("Failed to establish gRPC connection")
		return Fail
	}

	// Initialize the Nibiru client with the obtained gRPC connection.
	_, err = gonibi.NewNibiruClient(chainIDStr, grpcConn, rpcEndpointStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"chainID":     chainIDStr,
			"rpcEndpoint": rpcEndpointStr,
			"error":       err,
		}).Error("Failed to connect to Nibiru network")
		return Fail
	}

	// Log successful connection.
	logrus.WithFields(logrus.Fields{
		"chainID": chainIDStr,
	}).Info("Successfully connected to the Nibiru network")

	return Success
}

// GenerateRecoveryPhrase creates a new BIP39 mnemonic recovery phrase.
// It returns a pointer to a C string containing the mnemonic, or an empty string on failure.
//
//export GenerateRecoveryPhrase
func GenerateRecoveryPhrase() *C.char {
	// Log the start of the recovery phrase generation process.
	logrus.Info("Starting to generate a recovery phrase")

	// Define the entropy size for generating the mnemonic.
	const mnemonicEntropySize = 256

	// Generate entropy for the mnemonic.
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate entropy for recovery phrase")
		return C.CString("")
	}

	// Create a mnemonic from the generated entropy.
	phrase, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		logrus.WithError(err).Error("Failed to create mnemonic from entropy")
		return C.CString("")
	}

	// Log the successful generation of the recovery phrase.
	logrus.WithField("phrase", phrase).Info("Recovery phrase generated successfully")

	// Allocate memory for the phrase in C and return the pointer.
	cPhrase := C.CString(phrase)
	return cPhrase
}

// ToCKeyInfo converts KeyInfo to its C representation.
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

// CreateAccount creates a new blockchain account using the provided mnemonic, key name, and passphrase.
// It returns Success if the account creation is successful, otherwise Fail.
//
//export CreateAccount
func CreateAccount(keyName *C.char, mnemonic *C.char, passphrase *C.char) C.int {
	keyNameStr := C.GoString(keyName)
	mnemonicStr := C.GoString(mnemonic)
	// passphraseStr := C.GoString(passphrase)

	// Log the attempt to create a new account with the specified key name.
	logrus.WithField("keyName", keyNameStr).Info("Attempting to create new account")

	// Create a new signer using the provided mnemonic and key name.
	record, _, err := gonibi.CreateSigner(mnemonicStr, gosdk.Keyring, keyNameStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyName": keyNameStr,
			"error":   err,
		}).Error("Failed to create new account")
		return Fail
	}

	// Obtain the address of the newly created account.
	addr, err := record.GetAddress()
	if err != nil {
		logrus.WithError(err).Error("Failed to get address of the newly created account")
		return Fail
	}

	// Log the successful account creation.
	logrus.WithFields(logrus.Fields{
		"keyName": keyNameStr,
		"address": addr.String(),
	}).Info("Account created successfully")

	// Optionally print the list of all signers for debugging.
	PrintListSigners()

	return Success
}

// GetPrivKeyFromMnemonic retrieves a private key from a given mnemonic and key name,
// returning a pointer to the private key data in a C-compatible format.
//
//export GetPrivKeyFromMnemonic
func GetPrivKeyFromMnemonic(mnemonic *C.char, keyName *C.char) *C.uint8_t {
	mnemonicStr := C.GoString(mnemonic)
	keyNameStr := C.GoString(keyName)

	// Log the attempt to retrieve a private key using the mnemonic and key name.
	logrus.WithField("keyName", keyNameStr).Info("Attempting to retrieve private key from mnemonic")

	// Initialize the keyring.
	kring := gosdk.Keyring

	// Retrieve the private key from the mnemonic.
	privKey, _, err := gonibi.PrivKeyFromMnemonic(kring, mnemonicStr, keyNameStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyName": keyNameStr,
			"error":   err,
		}).Error("Failed to retrieve private key from mnemonic")
		return nil
	}

	// Log the successful retrieval of the private key.
	addressStr := privKey.PubKey().Address().String()
	logrus.WithFields(logrus.Fields{
		"keyName": keyNameStr,
		"address": addressStr,
	}).Info("Private key retrieved successfully")

	// Convert the private key bytes to a C-compatible format.
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
	PrintPayload("GetAddressFromKeyName", C.GoString(keyName))
	keyInfo, err := gosdk.Keyring.Key(C.GoString(keyName))
	if err != nil {
		logrus.Error("Failed to get address", err)
		return nil
	}
	addr, err := keyInfo.GetAddress()
	if err != nil {
		logrus.Error("Failed to get address", err)
		return nil
	}

	logrus.Info("Return Address: ", addr.String())

	return C.CString(addr.String())
}

//export ImportAccountFromMnemoic
func ImportAccountFromMnemoic(mnemonic *C.char, keyName *C.char) C.int {
	mnemonicStr := C.GoString(mnemonic)
	keyNameStr := C.GoString(keyName)
	PrintPayload("ImportAccountFromMnemoic", mnemonicStr, keyNameStr)
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
func ImportAccountFromPrivateKey(privateKey *C.uint8_t, keyName *C.char) C.int {
	PrintPayload("ImportAccountFromPrivateKey", C.GoString(keyName))
	// Decode the private key string from hex
	privKeyBytes := cUint8ToGoSlice(privateKey)
	if privKeyBytes == nil {
		logrus.Error("Can not get private key")
	}

	// Create a PrivKey instance and assign the decoded bytes to its Key field
	privKey := secp256k1.PrivKey{
		Key: privKeyBytes,
	}
	// Create a keyring
	signer, err := gonibi.CreateSignerFromPrivKey(&privKey, C.GoString(keyName))
	if err != nil {
		return Fail
	}
	logrus.Info("Success to import account: ", signer.Name)
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
	PrintPayload("GetAccountByKeyName", C.GoString(keyName))
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
	logrus.Infof("Name: %s\nPubkey: %s\n address: %s", keyInfo.Name, keyInfo.PubKey, addr.String())
	return convertKeyInfo(keyInfo)
}

//export GetAccountByAddress
func GetAccountByAddress(addr *C.char) *C.KeyInfo {
	PrintPayload("GetAccountByAddress", C.GoString(addr))
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
	logrus.Debug("Call HasKeyByName")
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
	logrus.Debug("Call HasKeyByAddres")
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

//export TransferToken
func TransferToken(fromAddress, toAddress, denom *C.char, amount C.int) C.int {
	logrus.Info("Call TransferToken")
	PrintPayload("TransferToken", C.GoString(fromAddress), C.GoString(toAddress), C.GoString(denom), amount)
	// Convert C strings to Go strings
	fromStr := C.GoString(fromAddress)
	toStr := C.GoString(toAddress)
	denomStr := C.GoString(denom)
	PrintBaseAccountInfo(fromStr, toStr)

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
func ExecuteWasmContract(senderAddress, contractAddress, executeMsg, denom *C.char, amount C.int) *C.char {
	PrintPayload("ExecuteWasmContract", C.GoString(senderAddress), C.GoString(contractAddress), C.GoString(executeMsg), C.GoString(denom), amount)
	// Convert C types to Go types
	fromStr := C.GoString(senderAddress)
	contractStr := C.GoString(contractAddress)
	msgStr := C.GoString(executeMsg)
	denomStr := C.GoString(denom)
	amountInt := sdk.NewInt(int64(amount))

	// Get the sender's address
	from, err := sdk.AccAddressFromBech32(fromStr)
	if err != nil {
		logrus.Error("Failed to parse sender address:", err)
		return nil
	}

	// Get the contract address
	contract, err := sdk.AccAddressFromBech32(contractStr)
	if err != nil {
		logrus.Error("Failed to parse contract address:", err)
		return nil
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
		return nil
	}

	logrus.Info("Response: ", string(responseMsg.String()))

	return C.CString(responseMsg.TxHash)
}

//export QueryWasmContract
func QueryWasmContract(contractAddress, queryMsg *C.char) *C.char {
	PrintPayload("QueryWasmContract", C.GoString(contractAddress), C.GoString(queryMsg))
	// Convert C types to Go types
	contractStr := C.GoString(contractAddress)
	msgStr := C.GoString(queryMsg)

	// Get the contract address
	contract, err := sdk.AccAddressFromBech32(contractStr)
	if err != nil {
		logrus.Error("Failed to parse contract address:", err)
		return nil
	}

	// Create the Wasm execute message
	msgExe := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contract.String(),
		QueryData: []byte(msgStr),
	}

	responseMsg, err := wasmClient.SmartContractState(context.Background(), msgExe)

	if err != nil {
		logrus.Error("Error SmartContractState", err)
		return nil
	}

	logrus.Info("Response: ", string(responseMsg.String()))

	return C.CString(responseMsg.String())
}

// QueryTXHash retrieves the transaction details corresponding to a given transaction hash.
// It returns a pointer to a C string containing the transaction log, or nil on failure.
//
//export QueryTXHash
func QueryTXHash(txHash *C.char) *C.char {
	// Convert C string to Go string.
	txHashStr := C.GoString(txHash)

	// Log the attempt to decode and query transaction hash.
	logrus.WithField("txHash", txHashStr).Info("Attempting to query transaction")

	// Decode hex string to bytes.
	decodedBytes, err := hex.DecodeString(txHashStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"txHash": txHashStr,
			"error":  err,
		}).Error("Failed to decode transaction hash")
		return nil
	}

	// Query transaction details using the decoded bytes.
	resultTx, err := gosdk.CometRPC.Tx(context.Background(), decodedBytes, true)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"txHash": txHashStr,
			"error":  err,
		}).Error("Failed to retrieve transaction details")
		return nil
	}

	// Log the successful retrieval of transaction details.
	logrus.WithFields(logrus.Fields{
		"txHash": txHashStr,
		"result": resultTx.TxResult.Log,
	}).Info("Transaction retrieved successfully")

	// Convert the transaction log into a C string to return.
	cLog := C.CString(resultTx.TxResult.Log)
	return cLog
}
