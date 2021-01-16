// v0.7.5
// Author: DIEHL E.
// (C) Sony Pictures Entertainment, Jan 2021

package blockchain

import (
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	cDiscoveryKey = "DISCOVERY_AS_LOCALHOST"
	cVersion      = "0.6.1-8a"
)

var Logr *logrus.Logger

// Client is the structure handling the connection to the blockchain.
type Client struct {
	gw          *gateway.Gateway
	wallet      *gateway.Wallet
	network     *gateway.Network
	contract    *gateway.Contract
	initialized bool
	walletDir   string
	local       bool // temporary fix for Bug v1.0.0-beta3.0.20201006151309-9c426dcc5096
}

// NewClient creates a new Client for the configuration defined by file `configFile`.
//  If the user is not yet in the
// wallet, it attempts to populate the wallet.  `options` may overwrite the data
// provided by the configuration file.
func NewClient(configFile string, path string, options ...ClientOption) (*Client, error) {
	// collects the options
	clOpts := clientOptions{user: "", walletDir: "", log: "bc.log"} // default values
	for _, option := range options {
		option(&clOpts)
	}

	vi := viper.New()
	vi.SetConfigName(configFile)
	vi.AddConfigPath(path)

	cp := &Configuration{}
	err := cp.Load(vi)
	if err != nil && err != ErrNoVault {
		Logr.Fatalf("could not load the configuration from %s: %v", configFile, err)
		return nil, ErrWalletInitFailed
	}

	if clOpts.user != "" {
		// overwrites the potential User defined in `configFile`
		cp.User = clOpts.user
	}

	// if the options did not define the wallet dir, then set default.
	if clOpts.walletDir == "" {
		clOpts.walletDir = filepath.Join(cp.configDir, "wallet")
	}

	c := &Client{walletDir: clOpts.walletDir}
	err = c.init(cp)
	return c, err
}

// Close closes the Client.
func (c *Client) Close() {
	if c.initialized {
		c.gw.Close()
	}
}

// Invoke submits the transaction `fn` with the argumenst `args`.
func (c *Client) Invoke(fn string, args ...string) ([]byte, error) {
	if !c.initialized {
		return nil, ErrClientNotInitialized
	}
	return c.contract.SubmitTransaction(fn, args...)
}

// func (c *Client) Init(configFile string, channelID string, chaincodeID string, user string, credPath string) error {
// 	return c.init(configFile, channelID, chaincodeID, user, credPath)
// }

// Query submits a transactiion `fn`with the arguments `args`.
func (c *Client) Query(fn string, args ...string) ([]byte, error) {
	if !c.initialized {
		return nil, ErrClientNotInitialized
	}
	// if c.local {
	// 	return c.contract.SubmitTransaction(fn, args...)
	// }

	return c.contract.EvaluateTransaction(fn, args...)
}

// init setups the discovery conditions, initializes the wallet if needed,
// and sets up contract.
func (c *Client) init(cp *Configuration) error {
	c.initialized = false
	c.local = cp.gatewayNotLocal // temporary
	if !cp.gatewayNotLocal {
		os.Setenv(cDiscoveryKey, "true")
	} else {
		os.Setenv(cDiscoveryKey, "false")

	}

	Logr.Debugf("%s = %s", cDiscoveryKey, os.Getenv(cDiscoveryKey))
	var err error
	c.wallet, err = gateway.NewFileSystemWallet(c.walletDir)
	if err != nil {
		Logr.Errorf("Failed to create wallet: %v", err)
		return err
	}

	if !c.wallet.Exists(cp.User) {
		// err = populateWallet(c.wallet, cp)
		// if err != nil {
		// 	Logr.Errorf("Failed to populate wallet contents: %v", err)
		// 	return err
		// }
	}
	Logr.Debug("wallet operational")
	Logr.Debugf("Connection file %s", filepath.Clean(cp.ConnectionFile))
	c.gw, err = gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(cp.ConnectionFile))),
		gateway.WithIdentity(c.wallet, cp.User),
	)
	if err != nil {
		Logr.Errorf("Failed to connect to gateway: %v", err)
		return err
	}
	Logr.Debug("gateway connected")
	c.network, err = c.gw.GetNetwork(cp.ChannelID)
	if err != nil {
		Logr.Errorf("Failed to get network: %v", err)
		os.Exit(1)
	}
	Logr.Debug("network acquired")
	c.contract = c.network.GetContract(cp.ChainCodeID)
	c.initialized = true
	return nil
}

// func populateWallet(wallet *gateway.Wallet, cp *Configuration) error {
// 	fs1 := &FabricSetup{}
// 	fs1.Configuration = *cp
// 	fmt.Println(fs1.Configuration)
// 	err := fs1.initSDK()
// 	// err := fs1.InitSDKLite(filepath.Base(cp.ConfigFile), filepath.Dir(cp.ConfigFile))
// 	if err != nil {
// 		Logr.Errorf("populateWallet: could not init SDK %s: %v", fs1.ConfigFile, err)
// 		return ErrSDKInitialized
// 	}
// 	if cp.UserPwd == "" {
// 		err = fs1.InitUser(cp.User)
// 	} else {
// 		// the enrollment secret is passed as params in the configuration.
// 		err = fs1.initUserWithSecret(cp.User, cp.UserPwd)
// 	}

// 	if err != nil {
// 		Logr.Errorf("populate wallet: cannot init the user %s: %v", cp.User, err)
// 		return ErrWalletInitFailed
// 	}
// 	fs1.CloseSDK()

// 	sk, err := fs1.getPrivateKeyName(cp.User)
// 	if err != nil {
// 		Logr.Errorf("populate wallet: private key %s not available", cp.User)
// 		return ErrWalletInitFailed
// 	}

// 	crt := cp.User + "@" + cp.PeerOrg + "-cert.pem"
// 	cert, err := ioutil.ReadFile(filepath.Join(cp.CredPath, crt))
// 	if err != nil {
// 		Logr.Errorf("cert file %s not available in %s", crt, cp.CredPath)
// 		return errors.New("cert file not available")
// 	}

// 	key, err := ioutil.ReadFile(filepath.Join(cp.CredPath, sk))
// 	if err != nil {
// 		return errors.New("private key file not available")
// 	}

// 	identity := gateway.NewX509Identity(cp.PeerOrg, string(cert), string(key))

// 	err = wallet.Put(cp.User, identity)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

type clientOptions struct {
	user      string
	walletDir string
	log       string
}

// ClientOption allows to parameterize the NewClient function.
type ClientOption func(opts *clientOptions)

// WithLog sets the lofg file to `name` instead of default "bc.log".
func WithLog(name string) ClientOption {
	return func(cp *clientOptions) {
		cp.log = name
	}
}

// WithUser sets the user of the client to `name` regardless of what was in the
// configuration file.
func WithUser(name string) ClientOption {
	return func(cp *clientOptions) {
		cp.user = name
	}
}

// WithWallet sets the directory of the wallet instead of the default "wallet"
// directory.
func WithWallet(dir string) ClientOption {
	return func(cp *clientOptions) {
		cp.walletDir = dir
	}
}

func init() {
	// logging.SetLevel("fabsdk/fab", logging.ERROR)
	initLogr("bc.log") // default logging file is "bc.log"
}

// initLogr initializes the logger Logr to write in the file defined by `fileName`.
//  It sets also logger "fabsdk/fab" to Error level.
func initLogr(fileName string) {
	Logr = logrus.New()

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {

		Logr.Out = file
	} else {
		Logr.Infof("Failed to log to file %s, using default stderr", fileName)

	}
	Logr.Level = logrus.InfoLevel
	Logr.Infof("Blockchain version %s", cVersion)
}
