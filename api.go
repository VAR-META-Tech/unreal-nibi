package main

import "C"

import (
	"log"
	"os"
	"unsafe"

	"github.com/NibiruChain/nibiru/x/common/testutil/cli"
	"github.com/Unique-Divine/gonibi"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/x/common/testutil/genesis"

	tmconfig "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

func main() {}

type NibiruClientService struct {
	suite.Suite

	gosdk    *gonibi.NibiruClient
	grpcConn *grpc.ClientConn
	cfg      *cli.Config
	network  *cli.Network
	val      *cli.Validator
}

//export NewNibiruClientService
func NewNibiruClientService() *C.char {
	// Here you would initialize your NibiruClientService.
	// For demonstration, we'll assume that the necessary components
	// can be initialized without any parameters for simplicity.

	service := &NibiruClientService{
		// Initialize the fields as necessary.
	}

	nibiru, err := CreateBlockchain()
	service.NoError(err)
	service.network = nibiru.Network
	service.cfg = nibiru.Cfg
	service.val = nibiru.Val
	service.grpcConn = nibiru.GrpcConn

	// Allocate enough memory in C for the service and return a pointer to it.
	// This memory must be freed from the C side when it's no longer needed.
	ptr := C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	*(*uintptr)(ptr) = uintptr(unsafe.Pointer(service))
	return (*C.char)(ptr)
}

//export RPCEndpoint
func RPCEndpoint(clientServicePtr *C.char) *C.char {
	// Convert the pointer back to a Go pointer.
	servicePtr := *(**NibiruClientService)(unsafe.Pointer(clientServicePtr))
	// Make sure that the RPCAddress is initialized before using it.
	if servicePtr.val != nil {
		return C.CString(servicePtr.val.RPCAddress)
	}
	return nil
}

type Blockchain struct {
	GrpcConn *grpc.ClientConn
	Cfg      *cli.Config
	Network  *cli.Network
	Val      *cli.Validator
}

type Logger interface {
	Log(v ...interface{})
	Logf(format string, args ...interface{})
}

type LogAdapter struct {
	*log.Logger
}

func (adapter LogAdapter) Log(v ...interface{}) {
	adapter.Print(v...)
}

func (adapter LogAdapter) Logf(format string, args ...interface{}) {
	adapter.Printf(format, args...)
}

// Support function
func CreateBlockchain() (nibiru Blockchain, err error) {
	gonibi.EnsureNibiruPrefix()
	encConfig := app.MakeEncodingConfig()
	genState := genesis.NewTestGenesisState(encConfig)
	cliCfg := cli.BuildNetworkConfig(genState)
	cfg := &cliCfg
	cfg.NumValidators = 1

	logger := log.New(os.Stdout, "prefix: ", log.LstdFlags)
	adapter := LogAdapter{logger}
	adapter.Log("TempDir" + os.TempDir())
	network, err := cli.New(adapter,
		os.TempDir(),
		*cfg,
	)
	if err != nil {
		return nibiru, err
	}
	err = network.WaitForNextBlock()
	if err != nil {
		return nibiru, err
	}

	val := network.Validators[0]
	AbsorbServerConfig(cfg, val.AppConfig)
	AbsorbTmConfig(cfg, val.Ctx.Config)

	grpcConn, err := ConnectGrpcToVal(val)
	if err != nil {
		return nibiru, err
	}
	return Blockchain{
		GrpcConn: grpcConn,
		Cfg:      cfg,
		Network:  network,
		Val:      val,
	}, err
}

func ConnectGrpcToVal(val *cli.Validator) (*grpc.ClientConn, error) {
	grpcUrl := val.AppConfig.GRPC.Address
	return gonibi.GetGRPCConnection(
		grpcUrl, true, 5,
	)
}

func AbsorbServerConfig(
	cfg *cli.Config, srvCfg *serverconfig.Config,
) *cli.Config {
	cfg.GRPCAddress = srvCfg.GRPC.Address
	cfg.APIAddress = srvCfg.API.Address
	return cfg
}

func AbsorbTmConfig(
	cfg *cli.Config, tmCfg *tmconfig.Config,
) *cli.Config {
	cfg.RPCAddress = tmCfg.RPC.ListenAddress
	return cfg
}

func (chain *Blockchain) TmRpcEndpoint() string {
	return chain.Val.RPCAddress
}
