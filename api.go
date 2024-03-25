package main

import "C"

import (
	"fmt"
	"unsafe"

	"github.com/Unique-Divine/gonibi"
)

// C representation of the NibiruClient struct
type C_NibiruClient struct {
	ChainId          *C.char
	Keyring          *C.char        // Assuming Keyring can be serialized
	EncCfg           *C.char        // Assuming EncodingConfig can be serialized
	Querier          *C.char        // Assuming Querier can be serialized
	CometRPC         *C.char        // Assuming cmtrpcclient.Client can be serialized
	AccountRetriever *C.char        // Assuming authtypes.AccountRetriever can be serialized
	GrpcClient       unsafe.Pointer // (See explanation below)
}

func main() {}

const (
	Success C.int = 0
	Fail          = 1
)

//export NewNibiruClient
func NewNibiruClient() C.int {
	fmt.Printf("Call NewNibiruClient")
	grpcConn, err := gonibi.GetGRPCConnection(gonibi.DefaultNetworkInfo.GrpcEndpoint, true, 2)
	if err != nil {
		fmt.Printf("GetGRPCConnection error: " + err.Error())
		return Fail
	}
	gosdk, err := gonibi.NewNibiruClient("nibiru-localnet-0", grpcConn, gonibi.DefaultNetworkInfo.TmRpcEndpoint)
	fmt.Printf("Connected to " + gosdk.ChainId)
	if err != nil {
		fmt.Printf("NewNibiruClient error: " + err.Error())
		return Fail
	}

	return Success
}
