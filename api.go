package main

import "C"

import (
	"github.com/Unique-Divine/gonibi"
	"github.com/cosmos/go-bip39"
	"github.com/sirupsen/logrus"
)

func main() {

}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
}

const (
	Success = 0
	Fail    = 1
)

// get keyring by memoniic

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
func CreateAccount(mnemonic *C.char) C.int {
	// Convert C strings to Go strings
	mnemonicStr := C.GoString(mnemonic)
	// Create a keyring
	kring := gonibi.NewKeyring()
	keyName := ""
	signer, privKey, err := gonibi.CreateSigner(mnemonicStr, kring, keyName)
	logrus.Print("signer, privKey", signer, privKey)
	if err != nil {
		return Fail
	}
	return Success
}

// kring keyring.Keyring, mnemonic string, keyName string,
//
//export GetPrivKeyFromMnemonic
func GetPrivKeyFromMnemonic(mnemoic *C.char) C.int {
	kring := gonibi.NewKeyring()
	keyName := ""
	privKey, _, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), keyName)
	logrus.Print("Private key:", privKey)
	if err != nil {
		return Fail
	}
	return Success
}

// kring keyring.Keyring, mnemonic string, keyName string,
//
//export GetAddressFromMnemonic
func GetAddressFromMnemonic(mnemoic *C.char) C.int {
	kring := gonibi.NewKeyring()
	keyName := ""
	_, addr, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), keyName)
	logrus.Print("Address:", addr)
	if err != nil {
		return Fail
	}
	return Success
}

// keyring *C.char, privateKey *C.char, keyName *C.char

//export AddSignerToKeyring
func AddSignerToKeyring(mnemoic *C.char) C.int {
	kring := gonibi.NewKeyring()
	keyName := ""
	privKey, _, err := gonibi.PrivKeyFromMnemonic(kring, C.GoString(mnemoic), keyName)
	if err != nil {
		return Fail
	}
	if err := gonibi.AddSignerToKeyring(kring, privKey, keyName); err != nil {
		return Fail
	}
	return Success
}
