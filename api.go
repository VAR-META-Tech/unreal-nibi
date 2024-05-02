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
		logrus.WithError(err).Error("Failed to initialize gRPC connection with endpoint ", networkInfo.GrpcEndpoint)
	} else {
		logrus.WithField("endpoint", networkInfo.GrpcEndpoint).Info("gRPC connection established successfully")
	}

	// Initialize the Nibiru client with the obtained gRPC connection.
	gosdk, err = gonibi.NewNibiruClient(networkInfo.ChainID, grpcConn, networkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.WithError(err).Error("Failed to initialize Nibiru client for chain ID ", networkInfo.ChainID)
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
	grpcConn, err := gonibi.GetGRPCConnection(grpcEndpointStr, false, 2)
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
func CreateAccount(keyname *C.char, mnemonic *C.char, passphrase *C.char) C.int {
	keynameStr := C.GoString(keyname)
	mnemonicStr := C.GoString(mnemonic)
	passphraseStr := C.GoString(passphrase)
	algo := hd.Secp256k1

	// Log the attempt to create a new account with the specified key name.
	logrus.WithField("keyname", keynameStr).Info("Attempting to create new account")

	// Create a new signer using the provided mnemonic and key name.
	record, err := gosdk.Keyring.NewAccount(keynameStr, mnemonicStr, passphraseStr, sdk.GetConfig().GetFullBIP44Path(), algo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyname": keynameStr,
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
		"keyName": keynameStr,
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

// GetAddressFromKeyName retrieves the blockchain address associated with a given key name.
// It returns the address as a C string or nil if an error occurs.
//
//export GetAddressFromKeyName
func GetAddressFromKeyName(keyName *C.char) *C.char {
	keyNameStr := C.GoString(keyName)
	PrintPayload("GetAddressFromKeyName", keyNameStr)

	// Retrieve the key information from the keyring using the key name.
	keyInfo, err := gosdk.Keyring.Key(keyNameStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyName": keyNameStr,
			"error":   err,
		}).Error("Failed to retrieve key from keyring")
		return nil
	}

	// Retrieve the address from the key information.
	address, err := keyInfo.GetAddress()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyName": keyNameStr,
			"error":   err,
		}).Error("Failed to get address from key info")
		return nil
	}

	// Log the retrieved address.
	logrus.WithFields(logrus.Fields{
		"keyName": keyNameStr,
		"address": address.String(),
	}).Info("Successfully retrieved address from key name")

	// Return the address as a C string.
	return C.CString(address.String())
}

// ImportAccountFromPrivateKey imports an account using a private key and associates it with a given key name.
// It returns Success if the import is successful, otherwise Fail.
//
//export ImportAccountFromPrivateKey
func ImportAccountFromPrivateKey(privateKey *C.uint8_t, keyName *C.char) C.int {
	keyNameStr := C.GoString(keyName)
	PrintPayload("ImportAccountFromPrivateKey", keyNameStr)

	// Convert C.uint8_t pointer to a Go byte slice
	privKeyBytes := cUint8ToGoSlice(privateKey)
	if privKeyBytes == nil {
		logrus.Error("Failed to convert private key from C.uint8_t to Go slice")
		return Fail
	}

	// Create a PrivKey instance with the private key bytes
	privKey := secp256k1.PrivKey{Key: privKeyBytes}

	// Attempt to create a signer from the private key
	signer, err := gonibi.CreateSignerFromPrivKey(&privKey, keyNameStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyName": keyNameStr,
			"error":   err,
		}).Error("Failed to create signer from private key")
		return Fail
	}

	logrus.WithField("keyName", signer.Name).Info("Successfully imported account")
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

// GetAccountByKeyName retrieves account information by key name and returns it as a C.KeyInfo struct.
//
//export GetAccountByKeyName
func GetAccountByKeyName(keyName *C.char) *C.KeyInfo {
	keyNameStr := C.GoString(keyName)
	PrintPayload("GetAccountByKeyName", keyNameStr)

	// Retrieve the key information from the keyring using the key name.
	keyInfo, err := gosdk.Keyring.Key(keyNameStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyName": keyNameStr,
			"error":   err,
		}).Error("Failed to retrieve account by key name")
		return nil
	}

	// Attempt to get the address from the key info.
	addr, err := keyInfo.GetAddress()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"keyName": keyNameStr,
			"error":   err,
		}).Error("Failed to get address from key info")
		return nil
	}

	// Log details about the account.
	logrus.WithFields(logrus.Fields{
		"keyName": keyNameStr,
		"pubKey":  keyInfo.PubKey,
		"address": addr.String(),
	}).Debug("Account details retrieved successfully")

	// Convert the key information to a C-compatible structure and return it.
	return convertKeyInfo(keyInfo)
}

// GetAccountByAddress retrieves account information based on the blockchain address and returns it as a C.KeyInfo struct.
//
//export GetAccountByAddress
func GetAccountByAddress(addr *C.char) *C.KeyInfo {
	addressStr := C.GoString(addr)
	PrintPayload("GetAccountByAddress", addressStr)

	// Convert the C string to a Go string and parse the address.
	address, err := sdk.AccAddressFromBech32(addressStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"address": addressStr,
			"error":   err,
		}).Error("Failed to parse account address")
		return nil
	}

	logrus.Printf("C address: %s, Nibiru address: %s", addressStr, address)

	// Retrieve the key information from the keyring by address.
	keyInfo, err := gosdk.Keyring.KeyByAddress(address)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"address": address.String(),
			"error":   err,
		}).Error("Failed to retrieve account by address")
		return nil
	}

	// Convert the key information to a C-compatible struct and return it.
	return convertKeyInfo(keyInfo)
}

// HasKeyByName checks if a key with the specified name exists in the keyring.
// It returns Success if the key exists, otherwise Fail.
//
//export HasKeyByName
func HasKeyByName(name *C.char) C.int {
	logrus.Debug("Checking for key by name")

	// Convert the C string to a Go string.
	keyName := C.GoString(name)

	// Attempt to retrieve the key from the keyring.
	_, err := gosdk.Keyring.Key(keyName)
	if err != nil {
		// Logging the error along with the key name for better context.
		logrus.WithFields(logrus.Fields{
			"keyName": keyName,
			"error":   err,
		}).Error("Failed to find key by name")
		return Fail
	}

	// If the key retrieval is successful, log the success and return Success.
	logrus.WithField("keyName", keyName).Debug("Key found")
	return Success
}

// HasKeyByAddress checks if a key corresponding to the given address exists in the keyring.
// It returns Success if the key exists, otherwise Fail.
//
//export HasKeyByAddress
func HasKeyByAddress(addr *C.char) C.int {
	logrus.Debug("Checking for key by address")

	// Convert C string to Go string and parse the address.
	addressStr := C.GoString(addr)
	address, err := sdk.AccAddressFromBech32(addressStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"address": addressStr,
			"error":   err,
		}).Error("Invalid address format")
		return Fail
	}

	// Check for the key in the keyring by address.
	a, err := gosdk.Keyring.KeyByAddress(address)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"address": address.String(),
			"error":   err,
		}).Error("Failed to retrieve key by address")
		return Fail
	}

	// If a key is found, log the key name and return Success.
	if a != nil {
		logrus.WithField("keyName", a.Name).Debug("Key found")
		return Success
	}

	// If no key is found, return Fail.
	return Fail
}

// PrintListSigners logs the list of all signers in the keyring.
func PrintListSigners() {
	logrus.Debug("Attempting to retrieve list of accounts from keyring")

	// Retrieve the list of signers from the keyring
	signers, err := gosdk.Keyring.List()
	if err != nil {
		logrus.WithError(err).Debug("Failed to retrieve signers from keyring")
		return
	}

	// Log each signer's details
	for _, signer := range signers {
		addr, err := signer.GetAddress()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"signerName": signer.Name,
				"error":      err,
			}).Error("Failed to get address for signer")
			continue
		}
		logrus.WithFields(logrus.Fields{
			"signerName": signer.Name,
			"address":    addr.String(),
		}).Info("Signer details")
	}
}

// DeleteAccount removes an account from the keyring based on the given key name.
//
//export DeleteAccount
func DeleteAccount(keyName *C.char, password *C.char) C.int {
	// Log the attempt to delete an account
	logrus.Debug("Attempting to delete account with key name")

	// Convert C char to Go string and attempt to delete the account
	keyNameStr := C.GoString(keyName)
	err := gosdk.Keyring.Delete(keyNameStr)
	if err != nil {
		// Log failure with error
		logrus.WithField("keyName", keyNameStr).WithError(err).Debug("Failed to delete account")
		return Fail
	}

	// Log success and optionally print the list of remaining signers
	logrus.WithField("keyName", keyNameStr).Debug("Account successfully deleted")
	PrintListSigners()
	return Success
}

// TransferToken transfers specified amount of tokens from one address to another.
// It returns Success if the transaction is successful, otherwise Fail.
//
//export TransferToken
func TransferToken(fromAddress, toAddress, denom *C.char, amount C.int) C.int {
	logrus.Info("Initiating token transfer")
	// Convert C strings to Go strings
	fromStr := C.GoString(fromAddress)
	toStr := C.GoString(toAddress)
	denomStr := C.GoString(denom)

	// Log the initiation of the transfer with relevant data
	logrus.WithFields(logrus.Fields{
		"from":   fromStr,
		"to":     toStr,
		"denom":  denomStr,
		"amount": amount,
	}).Info("Transfer details")

	// Print account information before the transaction
	PrintBaseAccountInfo(fromStr, toStr)

	// Parse the sender's address
	from, err := sdk.AccAddressFromBech32(fromStr)
	if err != nil {
		logrus.WithError(err).WithField("fromAddress", fromStr).Error("Failed to parse sender address")
		return Fail
	}

	// Parse the recipient's address
	to, err := sdk.AccAddressFromBech32(toStr)
	if err != nil {
		logrus.WithError(err).WithField("toAddress", toStr).Error("Failed to parse recipient address")
		return Fail
	}

	// Create a coin with the specified denomination and amount
	coins := sdk.NewCoins(sdk.NewCoin(denomStr, sdk.NewInt(int64(amount))))

	// Create a MsgSend message to transfer tokens
	msgSend := banktypes.NewMsgSend(from, to, coins)

	// Broadcast the transaction to the blockchain network
	_, err = gosdk.BroadcastMsgs(from, msgSend)
	if err != nil {
		logrus.WithError(err).Error("Failed to broadcast token transfer message")
		return Fail
	}

	// Print account information after the transaction
	defer PrintBaseAccountInfo(fromStr, toStr)

	logrus.WithFields(logrus.Fields{
		"from": fromStr,
		"to":   toStr,
	}).Info("Token transfer executed successfully")

	return Success
}

// ExecuteWasmContract executes a smart contract on the blockchain using the specified parameters.
// It returns a pointer to a C string containing the transaction hash, or nil if an error occurs.
//
//export ExecuteWasmContract
func ExecuteWasmContract(senderAddress, contractAddress, executeMsg, denom *C.char, amount C.int) *C.char {
	// Log the incoming payload for debug purposes.
	PrintPayload("ExecuteWasmContract", C.GoString(senderAddress), C.GoString(contractAddress), C.GoString(executeMsg), C.GoString(denom), amount)

	// Convert C types to Go types.
	fromStr := C.GoString(senderAddress)
	contractStr := C.GoString(contractAddress)
	msgStr := C.GoString(executeMsg)
	denomStr := C.GoString(denom)
	amountInt := sdk.NewInt(int64(amount))

	// Parse sender's address.
	from, err := sdk.AccAddressFromBech32(fromStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"senderAddress": fromStr,
			"error":         err,
		}).Error("Failed to parse sender address")
		return nil
	}

	// Parse contract address.
	contract, err := sdk.AccAddressFromBech32(contractStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"contractAddress": contractStr,
			"error":           err,
		}).Error("Failed to parse contract address")
		return nil
	}

	// Create the coins to send with the message.
	coins := sdk.NewCoins(sdk.NewCoin(denomStr, amountInt))

	// Construct the Wasm execute message.
	msgExe := &wasmtypes.MsgExecuteContract{
		Sender:   from.String(),
		Contract: contract.String(),
		Msg:      []byte(msgStr),
		Funds:    coins,
	}

	// Broadcast the transaction to the blockchain network.
	responseMsg, err := gosdk.BroadcastMsgs(from, msgExe)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"sender":   fromStr,
			"contract": contractStr,
			"error":    err,
		}).Error("Failed to broadcast execute contract message")
		return nil
	}

	// Log the response transaction hash.
	txHash := responseMsg.TxHash
	logrus.WithFields(logrus.Fields{
		"txHash": txHash,
	}).Info("Executed contract successfully")

	// Return the transaction hash as a C string.
	return C.CString(txHash)
}

// QueryWasmContract queries the state of a Wasm smart contract using a contract address and a query message.
// It returns a pointer to a C string containing the query response or nil if an error occurs.
//
//export QueryWasmContract
func QueryWasmContract(contractAddress, queryMsg *C.char) *C.char {
	// Log the payload for debugging purposes.
	PrintPayload("QueryWasmContract", C.GoString(contractAddress), C.GoString(queryMsg))

	// Convert C types to Go types for further processing.
	contractStr := C.GoString(contractAddress)
	msgStr := C.GoString(queryMsg)

	// Parse the contract address.
	contract, err := sdk.AccAddressFromBech32(contractStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"contractAddress": contractStr,
			"error":           err,
		}).Error("Failed to parse contract address")
		return nil
	}

	// Create the Wasm execute message.
	msgExe := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contract.String(),
		QueryData: []byte(msgStr),
	}

	// Perform the smart contract state query.
	responseMsg, err := wasmClient.SmartContractState(context.Background(), msgExe)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"contractAddress": contractStr,
			"query":           msgStr,
			"error":           err,
		}).Error("Failed to query smart contract state")
		return nil
	}

	// Convert the response to a string and log it.
	responseStr := responseMsg.String()
	logrus.WithFields(logrus.Fields{
		"contractAddress": contractStr,
		"response":        responseStr,
	}).Info("Smart contract state queried successfully")

	// Return the response as a C string.
	return C.CString(responseStr)
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
