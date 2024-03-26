package main

import "C"

import (
	"reflect"

	"github.com/Unique-Divine/gonibi"
	"github.com/cosmos/go-bip39"
	"github.com/sirupsen/logrus"
)

func main() {

}

// Declare a global variable to hold the gosdk instance
var gosdk gonibi.NibiruClient

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	// Initialize the gosdk variable
	// grpcConn, err := gonibi.GetGRPCConnection(gonibi.DefaultNetworkInfo.GrpcEndpoint, true, 2)
	// if err != nil {
	//     logrus.Fatalf("Failed to initialize gosdk: %s", err)
	// }
	// gosdk, err = gonibi.NewNibiruClient("nibiru-localnet-0", grpcConn, gonibi.DefaultNetworkInfo.TmRpcEndpoint)
	// if err != nil {
	//     logrus.Fatalf("Failed to initialize gosdk: %s", err)
	// }
	// logrus.Println("[init] gosdk initialized")
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
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

//export CreateAccount
func CreateAccount(keyName *C.char, mnemonic *C.char) C.int {
	logrus.Debug("Creating Account")
	mnemonicStr := C.GoString(mnemonic)

	// Create a keyring
	kring := gonibi.NewKeyring()
	signer, privKey, err := gonibi.CreateSigner(mnemonicStr, kring, C.GoString(keyName))
	logrus.Println("signer, privKey", signer, privKey)
	if err != nil {
		return Fail
	}
	return Success
}

//export GetPrivKeyFromMnemonic
func GetPrivKeyFromMnemonic(mnemoic *C.char, keyName *C.char) C.int {
	logrus.Debug("Call GetPrivKeyFromMnemonic")
	kring := gonibi.NewKeyring()
	privKey, _, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), C.GoString(keyName))
	logrus.Println("privkey: ", privKey, reflect.TypeOf(privKey))
	if err != nil {
		return Fail
	}
	return Success
}

//export GetAddressFromMnemonic
func GetAddressFromMnemonic(mnemoic *C.char, keyName *C.char) C.int {
	logrus.Println("Call GetAddressFromMnemonic")
	kring := gonibi.NewKeyring()
	_, addr, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), C.GoString(keyName))
	logrus.Println("Address:", addr)
	if err != nil {
		return Fail
	}
	return Success
}

//export AddSignerToKeyring
func AddSignerToKeyring(mnemoic *C.char, keyName *C.char) C.int {
	logrus.Debug("Call AddSignerToKeyring")
	kring := gonibi.NewKeyring()
	privKey, _, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), C.GoString(keyName))
	if err != nil {
		logrus.Debug("Failed to get private key", err)
		return Fail
	}
	if err := gonibi.AddSignerToKeyring(kring, privKey, C.GoString(keyName)); err != nil {
		logrus.Debug("Failed to add signer to keyring", err)
		return Fail
	}
	return Success
}

//export ImportAccount
func ImportAccount(mnemonic *C.char, privateKey *C.char, keyName *C.char) C.int {
	logrus.Debug("Import Account")
	mnemonicStr := C.GoString(mnemonic)
	// Create a keyring
	kring := gonibi.NewKeyring()
	signer, privKey, err := gonibi.CreateSigner(mnemonicStr, kring, C.GoString(keyName))
	logrus.Println("signer, privKey", signer, privKey)
	if err != nil {
		return Fail
	}
	return Success
}

//export DeleteAccount
func DeleteAccount(keyName *C.char, password *C.char) C.int {
	logrus.Debug("Call DeleteAccount")
	kring := gonibi.NewKeyring()
	if err := kring.Delete(C.GoString(keyName)); err != nil {
		logrus.Debug("Error:", err)
		return Fail
	}
	return Success
}
