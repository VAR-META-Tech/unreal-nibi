package main

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/nibiru/x/common/testutil"
	"github.com/NibiruChain/nibiru/x/common/testutil/cli"
	"github.com/NibiruChain/nibiru/x/common/testutil/genesis"
	tmconfig "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	"github.com/Unique-Divine/gonibi"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// --------------------------------------------------
// NibiruClientSuite
// --------------------------------------------------

var _ suite.SetupAllSuite = (*NibiruClientSuite)(nil)

type NibiruClientSuite struct {
	suite.Suite

	gosdk    *gonibi.NibiruClient
	grpcConn *grpc.ClientConn
	cfg      *cli.Config
	network  *cli.Network
	val      *cli.Validator
}

type Blockchain struct {
	GrpcConn *grpc.ClientConn
	Cfg      *cli.Config
	Network  *cli.Network
	Val      *cli.Validator
}

func CreateBlockchain(t *testing.T) (nibiru Blockchain, err error) {
	gonibi.EnsureNibiruPrefix()
	encConfig := app.MakeEncodingConfig()
	genState := genesis.NewTestGenesisState(encConfig)
	cliCfg := cli.BuildNetworkConfig(genState)
	cfg := &cliCfg
	cfg.NumValidators = 1

	network, err := cli.New(
		t,
		t.TempDir(),
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

func TestNibiruClientTestSuite_RunAll(t *testing.T) {
	suite.Run(t, new(NibiruClientSuite))
}

func (s *NibiruClientSuite) RPCEndpoint() string {
	return s.val.RPCAddress
}

// SetupSuite implements the suite.SetupAllSuite interface. This function runs
// prior to all of the other tests in the suite.
func (s *NibiruClientSuite) SetupSuite() {
	nibiru, err := CreateBlockchain(s.T())
	s.NoError(err)
	s.network = nibiru.Network
	s.cfg = nibiru.Cfg
	s.val = nibiru.Val
	s.grpcConn = nibiru.GrpcConn
}

func ConnectGrpcToVal(val *cli.Validator) (*grpc.ClientConn, error) {
	grpcUrl := TestNetworkInfo.GrpcEndpoint
	// grpcUrl := val.AppConfig.GRPC.Address
	return gonibi.GetGRPCConnection(
		grpcUrl, true, 5,
	)
}

func (s *NibiruClientSuite) ConnectGrpc() {
	grpcConn, err := ConnectGrpcToVal(s.val)
	s.NoError(err)
	s.NotNil(grpcConn)
	s.grpcConn = grpcConn
}

func (s *NibiruClientSuite) TestNewQueryClient() {
	_, err := gonibi.NewQueryClient(s.grpcConn)
	s.NoError(err)
}

func (s *NibiruClientSuite) TestNewNibiruClient() {
	rpcEndpt := s.val.RPCAddress
	gosdk, err := gonibi.NewNibiruClient(s.cfg.ChainID, s.grpcConn, rpcEndpt)
	s.NoError(err)
	s.gosdk = &gosdk

	s.gosdk.Keyring = s.val.ClientCtx.Keyring
	s.T().Run("DoTestBroadcastMsgs", func(t *testing.T) {
		s.DoTestBroadcastMsgs()
	})
	s.T().Run("DoTestBroadcastMsgsGrpc", func(t *testing.T) {
		s.NoError(s.network.WaitForNextBlock())
		s.DoTestBroadcastMsgsGrpc()
	})
}

func (s *NibiruClientSuite) UsefulPrints() {
	tmCfgRootDir := s.val.Ctx.Config.RootDir
	fmt.Printf("tmCfgRootDir: %v\n", tmCfgRootDir)
	fmt.Printf("s.val.Dir: %v\n", s.val.Dir)
	fmt.Printf("s.val.ClientCtx.KeyringDir: %v\n", s.val.ClientCtx.KeyringDir)
}

func (s *NibiruClientSuite) AssertTxResponseSuccess(txResp *sdk.TxResponse) (txHashHex string) {
	s.NotNil(txResp)
	s.EqualValues(txResp.Code, 0)
	return txResp.TxHash
}
func (s *NibiruClientSuite) msgSendVars() (from, to sdk.AccAddress, amt sdk.Coins, msgSend sdk.Msg) {
	from = s.val.Address
	to = testutil.AccAddress()
	amt = sdk.NewCoins(sdk.NewInt64Coin(denoms.NIBI, 420))
	msgSend = banktypes.NewMsgSend(from, to, amt)
	return from, to, amt, msgSend
}

func (s *NibiruClientSuite) DoTestBroadcastMsgs() (txHashHex string) {
	from, _, _, msgSend := s.msgSendVars()
	txResp, err := s.gosdk.BroadcastMsgs(
		from, msgSend,
	)
	s.NoError(err)
	return s.AssertTxResponseSuccess(txResp)
}

func (s *NibiruClientSuite) DoTestBroadcastMsgsGrpc() (txHashHex string) {
	from, _, _, msgSend := s.msgSendVars()
	txResp, err := s.gosdk.BroadcastMsgsGrpc(
		from, msgSend,
	)
	s.NoError(err)
	txHashHex = s.AssertTxResponseSuccess(txResp)

	base := 10
	var txRespCode string = strconv.FormatUint(uint64(txResp.Code), base)
	s.EqualValuesf(txResp.Code, 0,
		"code: %v\nraw log: %s", txRespCode, txResp.RawLog)
	return txHashHex
}

func (s *NibiruClientSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// --------------------------------------------------
// NibiruClientSuite_NoNetwork
// --------------------------------------------------

type NibiruClientSuite_NoNetwork struct {
	suite.Suite
}

func TestNibiruClientSuite_NoNetwork_RunAll(t *testing.T) {
	suite.Run(t, new(NibiruClientSuite_NoNetwork))
}

func (s *NibiruClientSuite_NoNetwork) TestGetGrpcConnection_NoNetwork() {
	grpcConn, err := gonibi.GetGRPCConnection(
		gonibi.DefaultNetworkInfo.GrpcEndpoint, true, 2,
	)
	s.Error(err)
	s.Nil(grpcConn)

	_, err = gonibi.NewQueryClient(grpcConn)
	s.Error(err)
}
