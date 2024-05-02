package test

import (
	"testing"
	"time"

	"github.com/Unique-Divine/gonibi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

var _ suite.SetupAllSuite = (*MainTestSuite)(nil)

type MainTestSuite struct {
	suite.Suite
	network          NetworkInfo
	mockGRPCConn     *grpc.ClientConn
	mockNibiruClient gonibi.NibiruClient
}

func TestMainTestSuite_RunAll(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

func (s *MainTestSuite) SetupSuite() {
	s.network = LocalNetworkInfo
	mockGRPCConn, err := gonibi.GetGRPCConnection(s.network.GrpcEndpoint, true, 2)
	s.NoError(err)
	s.mockGRPCConn = mockGRPCConn
	mockNibiruClient, err := gonibi.NewNibiruClient(s.network.ChainID, s.mockGRPCConn, s.network.TmRpcEndpoint)
	s.NoError(err)
	s.mockNibiruClient = mockNibiruClient
}

func (s *MainTestSuite) InitClients() {
	err := InitClients()
	s.NoError(err)
}

// Test SwitchNetwork
func (s *MainTestSuite) SwitchNetwork() {
	check := SwitchNetwork("test")
	s.Equal(0, check)
}

func (s *MainTestSuite) TestNetworkInit() {

	s.T().Run("InintClients", func(t *testing.T) {
		s.InitClients()
	})
	s.T().Run("SwitchNetwork", func(t *testing.T) {
		s.SwitchNetwork()
	})
}

// QueryAccount Test
func (s *MainTestSuite) TestQueryAccount() {
	// create Account first
	phrase := GenerateRecoveryPhrase()
	CreateAccount("test_key", phrase, "pass")
	addr := GetAddressFromKeyName("test_key")
	acc, err := QueryAccount(addr)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), acc)
}

// Test GetAccountCoins
func (s *MainTestSuite) TestGetAccountCoins() {

	addr := "nibi1zaavvzxez0elundtn32qnk9lkm8kmcsz44g7xl" //admin addr
	coins, err := GetAccountCoins(addr)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), coins)
}

// Test GenerateRecoveryPhrase
func (s *MainTestSuite) GenerateRecoveryPhrase() {
	phrase := GenerateRecoveryPhrase()
	assert.NotNil(s.T(), phrase)
}

// Test CreateAccount
func (s *MainTestSuite) TestCreateAccount() {
	phrase := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
	check := CreateAccount("test_key", phrase, "pass")
	s.Equal(0, check)
}

// Test GetPrivateKeyFromMnemonic
func (s *MainTestSuite) TestGetPrivateKeyFromMnemonic() {
	phrase := GenerateRecoveryPhrase()
	privKey := GetPrivKeyFromMnemonic(phrase, "test_key")
	s.NotNil(privKey)
}

// Test GetAddressFromKeyName
func (s *MainTestSuite) TestGetAddressFromKeyName() {
	phrase := GenerateRecoveryPhrase()
	CreateAccount("test_key", phrase, "pass")
	addr := GetAddressFromKeyName("test_key")
	s.NotNil(addr)
}

// Test ImportAccountFromMnemoic
func (s *MainTestSuite) TestImportAccountFromMnemoic() {
	phrase := GenerateRecoveryPhrase()

	check := ImportAccountFromMnemoic(phrase, "test_key")
	s.Equal(Success, check)
}

// Test ImportAccountFromPrivateKey
func (s *MainTestSuite) TestImportAccountFromPrivateKey() {
	phrase := GenerateRecoveryPhrase()
	keyName := "TestKey"
	privKey := GetPrivKeyFromMnemonic(phrase, keyName)

	s.T().Run("ImportAccountFromPrivateKey_Success", func(t *testing.T) {
		result := ImportAccountFromPrivateKey(privKey, keyName)
		if result != Success {
			t.Errorf("Expected success, but got %d", result)
		}
	})

	s.T().Run("ImportAccountFromPrivateKey_Failure", func(t *testing.T) {
		// Test with nil private key
		result := ImportAccountFromPrivateKey(nil, keyName)
		if result != Fail {
			t.Errorf("Expected failure, but got %d", result)
		}
	})
}

// Test Get AccountByKeyName
func (s *MainTestSuite) TestGetAccountByKeyName() {
	phrase := GenerateRecoveryPhrase()
	check := CreateAccount("test_key", phrase, "pass")
	assert.Equal(s.T(), Success, check)
	acc := GetAccountByKeyName("test_key")
	s.Equal("test_key", acc.Name)
}

// Test GetAccountByAddress
func (s *MainTestSuite) TestGetAccountByAddress() {
	phrase := GenerateRecoveryPhrase()
	CreateAccount("test_key", phrase, "pass")
	addr := GetAddressFromKeyName("test_key")
	acc := GetAccountByAddress(addr)
	assert.NotNil(s.T(), acc)
}

// Get HasKeyByName
func (s *MainTestSuite) TestHasKeyByName() {
	phrase := GenerateRecoveryPhrase()
	CreateAccount("test_key", phrase, "pass")
	check := HasKeyByName("test_key")
	s.Equal(0, check)
}

// Test HasKeyByAddress
func (s *MainTestSuite) TestHasKeyByAddress() {
	phrase := GenerateRecoveryPhrase()
	CreateAccount("test_key", phrase, "pass")
	addr := GetAddressFromKeyName("test_key")
	check := HasKeyByAddress(addr)
	s.Equal(0, check)
}

// Test DeleteAccount
func (s *MainTestSuite) TestDeleteAccount() {
	phrase := GenerateRecoveryPhrase()
	CreateAccount("test_key", phrase, "pass")
	check := DeleteAccount("test_key", "pass")
	s.Equal(0, check)
}

func (s *MainTestSuite) TestTransferToken() {

	phrase := "napkin rigid magnet grass plastic spawn replace hobby tray eternal pupil olive pledge nasty animal base bitter climb guess analyst fat neglect zoo earn"
	adminPhases := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"
	// create account from these phrases first
	// then use the address to test the transfer token
	check1 := CreateAccount("admin", adminPhases, "pass")
	check2 := CreateAccount("test", phrase, "pass")
	assert.Equal(s.T(), 0, check1)
	assert.Equal(s.T(), 0, check2)

	adminAddr := GetAddressFromKeyName("admin")
	testAddr := GetAddressFromKeyName("test")
	assert.NotNil(s.T(), adminAddr)
	assert.NotNil(s.T(), testAddr)
	assert.NotEmpty(s.T(), adminAddr)
	assert.NotEmpty(s.T(), testAddr)
	denom := "unibi"
	amount := 75

	s.T().Run("TransferToken_Success", func(t *testing.T) {
		result := TransferToken(adminAddr, testAddr, denom, amount)
		assert.NotEmpty(t, result)
	})

	s.T().Run("TransferToken_InvalidFromAddress", func(t *testing.T) {
		// Test with an invalid fromAddress
		result := TransferToken("invalid_address", testAddr, denom, amount)
		assert.NotEmpty(t, result)
	})

	s.T().Run("TransferToken_InvalidToAddress", func(t *testing.T) {
		// Test with an invalid toAddress
		result := TransferToken(adminAddr, "invalid_address", denom, amount)
		assert.NotEmpty(t, result)
	})
}
func (s *MainTestSuite) TestExecuteAndQueryWasmContract() {
	adminPhases := "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host"

	check1 := CreateAccount("admin", adminPhases, "pass")
	assert.Equal(s.T(), 0, check1)

	adminAddr := GetAddressFromKeyName("admin")
	assert.NotNil(s.T(), adminAddr)

	contractAddr := "nibi1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3slkhcux"
	executeMsg := "{\"mint\": {\"token_id\": \"unique-nft-15\", \"owner\": \"nibi1zy7amen6h5e4whcta4ac656l0whsalzmnqrkc5\", \"token_uri\": \"https://metadata.com/nft1.json\"}}"
	s.T().Run("ExecuteWasmContract", func(t *testing.T) {
		txHash := ExecuteWasmContract(adminAddr, contractAddr, executeMsg, "unibi", 75)
		s.NotNil(txHash)
		s.NotEmpty(txHash)

		time.Sleep(3 * time.Second)
		txResult := QueryTXHash(txHash)
		s.NotNil(txResult)
		s.NotEmpty(txResult)
	})

	time.Sleep(3 * time.Second)

	s.T().Run("QueryWasmContract", func(t *testing.T) {
		queryMsg := "{\"owner_of\": {\"token_id\": \"unique-nft-15\", \"include_expired\": false}}"
		result := QueryWasmContract(contractAddr, queryMsg)
		s.NotNil(result)
		s.NotEmpty(result)
	})
}
