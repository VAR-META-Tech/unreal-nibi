package test

import (
	"testing"

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
	addr := "cosmos1qperwt9wrnkg5k9e5gzfgjppzpqhyav5j24d66"
	acc, err := QueryAccount(addr)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), acc)
}

// Test GetAccountCoins
func (s *MainTestSuite) TestGetAccountCoins() {
	addr := "cosmos1qperwt9wrnkg5k9e5gzfgjppzpqhyav5j24d66"
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
	phrase := GenerateRecoveryPhrase()
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
	check := ImportAccountFromMnemoic("test_key", phrase)
	s.Equal(0, check)
}

// Test ImportAccountFromPrivateKey
func (s *MainTestSuite) TestImportAccountFromPrivateKey() {
	phrase := GenerateRecoveryPhrase()
	privKey := GetPrivKeyFromMnemonic(phrase, "test_key")
	check := ImportAccountFromPrivateKey(privKey, "key_name")
	s.Equal(0, check)
}

func TestImportAccountFromPrivateKey(t *testing.T) {
	privateKey := []byte{0x01, 0x02, 0x03} // Replace with your private key bytes
	keyName := "TestKey"

	t.Run("ImportAccountFromPrivateKey_Success", func(t *testing.T) {
		result := ImportAccountFromPrivateKey(privateKey, keyName)
		if result != Success {
			t.Errorf("Expected success, but got %d", result)
		}
	})

	t.Run("ImportAccountFromPrivateKey_Failure", func(t *testing.T) {
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
	CreateAccount("test_key", phrase, "pass")
	acc := GetAccountByKeyName("test_key")
	assert.NotNil(s.T(), acc)
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
		assert.Equal(t, Success, result)
	})

	s.T().Run("TransferToken_InvalidFromAddress", func(t *testing.T) {
		// Test with an invalid fromAddress
		result := TransferToken("invalid_address", testAddr, denom, amount)
		assert.Equal(t, Fail, result)
	})

	s.T().Run("TransferToken_InvalidToAddress", func(t *testing.T) {
		// Test with an invalid toAddress
		result := TransferToken(adminAddr, "invalid_address", denom, amount)
		assert.Equal(t, Fail, result)
	})
}

func TestExecuteWasmContract(t *testing.T) {
	senderAddress := "cosmos1qperwt9wrnkg5k9e5gzfgjppzpqhyav5j24d66"
	contractAddress := "cosmos1qperwt9wrnkg5k9e5gzfgjppzpqhyav5j24d66"
	executeMsg := "execute"
	denom := "unibi"
	amount := 100

	result := ExecuteWasmContract(senderAddress, contractAddress, executeMsg, denom, amount)

	assert.NotEmpty(t, result)
	assert.NotEqual(t, "nil", result)
}
