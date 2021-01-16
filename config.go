// v0.2.3
// Author: DIEHL E.
// (C) Sony Pictures Entertainment, Nov 2020

package blockchain

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Configuration holds the inofRmation needed to set up the full connection
// wityh the blockchain.
type Configuration struct {
	// configDir is the directory that holds the configuration files.
	// It is the field [sdk.dir] in the app config.toml.
	configDir string
	// // ConfigFile contains the name of the configuration file for the SDK.
	ConfigFile string
	// PeerOrg is the MSPID of the peer organization.  It is auto-populated
	PeerOrg string
	// OrdererID is the name of the orderer organiztion.  Mandatory for full.
	OrdererID string
	// ChannelID is the name of the channel on which to operate.  Mandatory
	ChannelID string
	// ChainCodeID is the name of the chaincode within the channel.  Mandatory for fulll
	ChainCodeID string
	// ChannelConfig is the path to the channel TX descriptor.  Mandatory for full
	ChannelConfig string
	// ChaincodePath is the path to the chaincode to be packaged
	// If defined, then ChaincodePackage should be empty.
	ChaincodePath string
	// ChaincodePackage is the path to the file that contains the packaged
	// chaincode.  If defined, then ChaincodePath should be empty.
	ChaincodePackage string
	// ChaincodeVersion is the version of the chaincode. Mandatory for intsallation.
	ChaincodeVersion string
	// OrgAdmin is the peer name admin.  Mandatory for full
	OrgAdmin string
	// Org name.  It is auto-populated.
	OrgName string
	// connection configuration file used by gateway.
	ConnectionFile string
	// User's name>  Mandatory for Client
	User string
	// UserPWD is [optional] user's enrolment scret.  If not provided, it will be
	// enrolled.
	UserPwd string
	// CredPath is the path to the user's key store.  It is auto-populated.
	CredPath string
	// gatewayNotLocal is false if the gateway interacts without a local docker. It is
	// the field "notlocal".
	gatewayNotLocal bool

	sdkDefined     bool
	gatewayDefined bool
}

// Load retrieves the configuration information from the viper reader `vi`.
func (c *Configuration) Load(vi *viper.Viper) error {
	err := vi.ReadInConfig()
	if err != nil {
		Logr.Errorf("could not read config file: %v", err)
		return err
	}
	if vi.GetBool("debug") {
		Logr.Level = logrus.DebugLevel
		Logr.Info("Log in debug mode")
	} else {
		Logr.Level = logrus.InfoLevel
		Logr.Info("Log in info mode")
	}

	fn := vi.GetString("log")
	if fn != "" {
		// the configuration file proposed a different filename  than default one
		initLogr(fn)
	}

	c.sdkDefined = true
	if vi.GetString("sdk.ChannelID") == "" || vi.GetString("sdk.ChaincodeID") == "" {
		c.sdkDefined = false
	}
	c.gatewayDefined = true
	if vi.GetString("gateway.Connection") == "" {
		c.gatewayDefined = false
	}
	if !c.sdkDefined && !c.gatewayDefined {
		Logr.Fatal("configuration file misses important data in [sdk] and/or [gateway] sections")
		return ErrSDKFailed
	}

	// find the configuration directory
	// ------------
	if c.configDir == "" {
		// It means that the directory was not overwritten by the WithDir options
		c.configDir = vi.GetString("sdk.dir")
		if c.configDir == "" {
			c.configDir = vi.GetString("gateway.dir")
		}
		if c.configDir == "" {
			Logr.Fatal("configuration file misses the field dir.")
		}
	}

	if !filepath.IsAbs(c.configDir) {
		spath, _ := os.Executable()
		c.configDir = filepath.Join(spath, c.configDir)
	}

	// Treat the gateway data
	// -----
	if c.gatewayDefined {
		cfg := vi.GetString("gateway.Connection")
		if cfg == "" {
			// field not defined.
			cfg = "connection.yaml"
		}
		c.ConnectionFile = filepath.Join(c.configDir, cfg)
		c.User = vi.GetString("gateway.User")
		c.UserPwd = vi.GetString("gateway.UserPwd")
		c.ChannelID = vi.GetString("gateway.ChannelID")
		c.ChainCodeID = vi.GetString("gateway.ChaincodeID")
		c.gatewayNotLocal = vi.GetBool("gateway.notlocal")
	}

	if c.sdkDefined {

		// Definition of the Fabric SDK properties
		c.OrdererID = vi.GetString("sdk.OrdererID")

		// Channel parameters
		if c.ChannelID == "" {
			// in case it was not in the gateway section.
			c.ChannelID = vi.GetString("sdk.ChannelID")
		}

		c.ChannelConfig = vi.GetString("sdk.ChannelConfig")
		// Chaincode parameters
		if c.ChainCodeID == "" {
			// in case it was not in the gateway section
			c.ChainCodeID = vi.GetString("sdk.ChaincodeID")
		}

		c.ChaincodeVersion = vi.GetString("sdk.ChaincodeVersion")
		c.ChaincodePath = vi.GetString("sdk.ChaincodePath")
		c.ChaincodePackage = vi.GetString("sdk.ChaincodePackage")
		c.OrgAdmin = vi.GetString("sdk.OrgAdmin")

		cfg := vi.GetString("sdk.ConfigFile")
		if cfg == "" {
			// defualt value if not defined in the toml file.
			cfg = "config.yaml"
		}
		c.ConfigFile = filepath.Join(c.configDir, cfg)
		vaultFile := filepath.Join(c.configDir, "store", "store.key")
		Logr.WithFields(logrus.Fields{"configFile": c.ConfigFile, "key": vaultFile}).Debug("select files")
		// err := setVault(vaultFile, true)

		// load the next parameters from the SDK config file.
		vi2 := viper.New()
		vi2.SetConfigName(Strip(filepath.Base(c.ConfigFile), "yaml"))
		vi2.SetConfigType("yaml")
		vi2.AddConfigPath(filepath.Dir(c.ConfigFile))
		err1 := vi2.ReadInConfig()
		if err1 != nil {

			Logr.Errorf("could not read SDK configuration file %s: %v", c.ConfigFile, err1)

			return ErrSDKFailed
		}
		c.OrgName = vi2.GetString("client.organization")
		c.CredPath = vi2.GetString("client.credentialStore.path")
		c.PeerOrg = vi2.GetString("organizations." + c.OrgName + ".mspid")

		// return err
	}
	return nil
}

// // selectConfig selects the proper config file depending on the url type.
// // vi is the pointer to the viper that has been read.  Currently, it
// // supports 0, 1, 2, 3, 4 and -1. -1 is for testing.
// func selectConfig(vi *viper.Viper) (string, error) {
// 	spath, _ := os.Executable()

// 	// The assumption is the following directory structure
// 	// cmd/app1/  contains the exec of app1
// 	//     app2/
// 	//     config/ contains the fabric-sdk-go yaml file
// 	//     config/store contains the vault files
// 	configDir := filepath.Dir(filepath.Join(spath, "..")) // backward one level
// 	os.Setenv("FABRICPATH", configDir)
// 	sdir := filepath.Join(configDir, "config")

// 	choice := vi.GetInt("url.type")
// 	storeDir := filepath.Join(sdir, "store")
// 	var cfg string
// 	vf := "store.key"
// 	absolute := true // indicates whether tyhe path to the vault is absolute.  It is
// 	// relative if in test mode.
// 	switch choice {
// 	case 0:
// 		cfg = "config.yaml"

// 	case -1: //test case.  The config file is in testdata/config.yaml
// 		sdir = vi.GetString("testConfig")
// 		// assumes that for pkg test the structure is
// 		// pkg/blockhain/testdata/
// 		// cmd/config/store
// 		// storeDir = filepath.Join(sdir, "../../cmd/config/store")
// 		// Assumes that the structure is sdir/store/
// 		storeDir = filepath.Join(sdir, "store")
// 		Logr.Debugf("storedir->%s", storeDir)
// 		cfg = "config.yaml"

// 		// absolute = false
// 	case -2:
// 		// the sdk-yaml file name is defined by the value in the config file
// 		cfg = vi.GetString("sdk.ConfigFile")
// 	default:
// 		if choice < -2 {
// 			Logr.Error("not proper url.type value in config file")
// 			return "", ErrNotProperNetwork
// 		}
// 		cfg = fmt.Sprintf("configCloud%d.yaml", choice)
// 		vf = fmt.Sprintf("store%d.key", choice)
// 	}

// 	return filepath.Join(sdir, cfg), setVault(filepath.Join(storeDir, vf), absolute)
// }

// Strip removes the extension `ext` if present.  If there is no trailing
// '.', it is added.  It returns the file name without the extension if present.
// The extension cam be composed, i.e., ".xxx.yyy".
func Strip(name string, ext string) string {
	if ext == "" {
		return name
	}
	if ext[:1] != "." {
		ext = "." + ext
	}
	return strings.TrimSuffix(name, ext)
}
