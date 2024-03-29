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
	"unsafe"

	"github.com/Unique-Divine/gonibi"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/sirupsen/logrus"
)

func main() {

}

var gosdk gonibi.NibiruClient

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	grpcConn, err := gonibi.GetGRPCConnection(gonibi.DefaultNetworkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		logrus.Fatalf("Failed to initialize Nibiru client: %s", err)
	}

	gosdk, err = gonibi.NewNibiruClient("nibiru-localnet-0", grpcConn, gonibi.DefaultNetworkInfo.TmRpcEndpoint)
	if err != nil {
		logrus.Fatalf("Failed to initialize Nibiru client: %s", err)
	}

	logrus.Println("[init] Nibiru client initialized")
}

const (
	Success = 0
	Fail    = 1
)

//export NewNibiruClientDefault
func NewNibiruClientDefault() C.int {
	logrus.Println("Call [NewNibiruClientDefault]") // Use logrus instead of fmt.Println
	grpcConn, err := gonibi.GetGRPCConnection(gonibi.DefaultNetworkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		logrus.Println("[NewNibiruClientDefault] GetGRPCConnection error: " + err.Error())
		return Fail
	}

	gosdk, err := gonibi.NewNibiruClient("nibiru-localnet-0", grpcConn, gonibi.DefaultNetworkInfo.TmRpcEndpoint)
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
	logrus.Debug("Call Creating Account")
	algo := hd.Secp256k1
	// Create a keyring
	record, err := gosdk.Keyring.NewAccount(C.GoString(keyName), C.GoString(mnemonic), C.GoString(passphase), sdk.GetConfig().GetFullBIP44Path(), algo)

	logrus.Println("Account created:", record.String())
	if err != nil {
		logrus.Debug("Failed to create new account", err)
		return Fail
	}
	PrintListSigners()
	return Success
}

//export CreateAccountV2
func CreateAccountV2(keyName *C.char, mnemonic *C.char, passphase *C.char) C.int {
	logrus.Debug("Call Creating Account")
	// Create a keyring
	keyRing := gosdk.Keyring
	mnemonicStr := C.GoString(mnemonic)
	signer, _, err := gonibi.CreateSigner(mnemonicStr, keyRing, C.GoString(keyName))
	logrus.Println("Account Created:", signer.String())
	if err != nil {
		return Fail
	}
	return Success
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
	logrus.Info(" C.CString(privKey.Bytes())", privKey.Bytes())
	logrus.Info("Priv Pub String", privKey.PubKey().String())
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

//export GetAddressFromMnemonic
func GetAddressFromMnemonic(mnemoic *C.char, keyName *C.char) *C.char {
	logrus.Println("Call GetAddressFromMnemonic")
	kring := gosdk.Keyring
	_, addr, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), C.GoString(keyName))
	if err != nil {
		logrus.Debug("Failed to get address", err)
		return nil
	}
	return C.CString(addr.String())
}

//export ImportAccountFromMnemoic
func ImportAccountFromMnemoic(mnemonic *C.char, keyName *C.char) C.int {
	logrus.Debug("Import Account")
	mnemonicStr := C.GoString(mnemonic)
	// Create a keyring
	kring := gosdk.Keyring
	signer, _, err := gonibi.CreateSigner(mnemonicStr, kring, C.GoString(keyName))
	if err != nil {
		logrus.Debug("Failed to import account:", err)
		return Fail
	}
	logrus.Println("Susscess to import account:", signer.Name, signer.PubKey.String(), signer.GetLedger().String())
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

	logrus.Info("Priv Pub String", privKey.PubKey().String())
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

	logrus.Debug("Account find: ", keyInfo)
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
		logrus.Debug("Signer name: ", signer.Name, signer.GetType())
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
