// V 0.6.4
// Author: DIEHL E.
// (C) Sony Pictures Entertainment, Nov 2020

package blockchain

import (
	"encoding/hex"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/spf13/viper"
)

// FabricSetup implementation of the fabric-sdk-go interface
type FabricSetup struct {
	Configuration
	// initialized is true if it the SDK was initializes fully.  In that case, initializedLite
	// is also true.
	initialized     bool
	initializedLite bool         // is true if the SDK was initialized in lite mode.
	vi              *viper.Viper // holds the viper config file.
	currentUser     string       // holds the name of the current user of client.

	client        *channel.Client // used to query smart contracts
	resMgmtClient *resmgmt.Client // used to manage channels
	sdk           *fabsdk.FabricSDK
	event         *event.Client
}

// // Attribute is a typical Key/Value structure to define optional
// // attributes for a user.
// type Attribute struct {
// 	Key   string
// 	Value string
// }

// // CreateUser registers and enrolls a new user if it is not yet known by
// // fabric-ca.  name is the public ID of the new user.  roleName defines
// // the expected role in the blockchain ecosystem.  This role is used by the
// // ABAC of the chaincode to filter the authorized smart contracts.
// // `attr` defines additional attributes for the user certificate that
// // are chaincode-dependent.
// //
// // If the user already exists, then the error ErrUserAlreadyExist is returned.
// func (fs *FabricSetup) CreateUser(name string, roleName string,
// 	attr ...Attribute) error {

// 	// Creates the context
// 	ctxProvider2 := fs.sdk.Context(fabsdk.WithOrg(fs.OrgName))
// 	mspClient, err := msp.New(ctxProvider2)
// 	if err != nil {
// 		Logr.Fatalf("CreateNewUser: Failed to create new signing client: %v", err)
// 		return ErrCreateUser
// 	}
// 	Logr.Debug("CreateNewUser: created signing client")

// 	// checks whether the user is not already known
// 	_, err = mspClient.GetIdentity(name)
// 	if err == nil {
// 		Logr.Infof("%s already exist.  Skip the creation.", name)
// 		return ErrUserAlreadyExist
// 	}

// 	req := generateReq(name, roleName, attr...)
// 	secret, err := mspClient.Register(req)
// 	if err != nil {
// 		Logr.Errorf("could not register %s due to %v", name, err)
// 		return ErrCreateUser
// 	}

// 	Logr.Debugf("createUser %s with %s", name, secret)
// 	err = storeSecret(name, secret)
// 	if err != nil {
// 		Logr.Fatalf("could not store secret in vault: %v", err)
// 		return ErrCreateUser
// 	}
// 	err = mspClient.Enroll(name, msp.WithSecret(secret))
// 	if err != nil {
// 		Logr.Errorf("could not enroll %s due to %v", name, err)
// 		return ErrCreateUser
// 	}

// 	Logr.Infof("Created User %s", name)
// 	return nil
// }

// // CreateUserWithSecret registers and enrolls a new user `name` using the enrolment
// // secret `secret` if it is not yet known by
// // fabric-ca.  `roleName` defines
// // the expected role in the blockchain ecosystem.
// func (fs *FabricSetup) CreateUserWithSecret(name string, roleName string, secret string,
// 	attr ...Attribute) error {

// 	// Creates the context
// 	ctxProvider2 := fs.sdk.Context(fabsdk.WithOrg(fs.OrgName))
// 	mspClient, err := msp.New(ctxProvider2)
// 	if err != nil {
// 		Logr.Fatalf("CreateNewUserWithSecret: Failed to create new signing client: %v", err)
// 		return ErrCreateUser
// 	}
// 	Logr.Debug("CreateNewUserWithSecret: created signing client")

// 	// checks whether the user is not already known
// 	_, err = mspClient.GetIdentity(name)
// 	if err == nil {
// 		Logr.Infof("%s already exist.  Skip the creation.", name)
// 		return ErrUserAlreadyExist
// 	}

// 	req := generateReq(name, roleName, attr...)
// 	req.Secret = secret

// 	_, err = mspClient.Register(req)
// 	if err != nil {
// 		Logr.Errorf("CreateNewUserWithSecret: could not register %s due to %v", name, err)
// 		return ErrCreateUser
// 	}

// 	Logr.Debugf("createUser %s with %s", name, secret)
// 	err = storeSecret(name, secret)
// 	if err != nil {
// 		Logr.Fatalf("could not store secret in vault: %v", err)
// 		return ErrCreateUser
// 	}
// 	err = mspClient.Enroll(name, msp.WithSecret(secret))
// 	if err != nil {
// 		Logr.Errorf("could not enroll %s due to %v", name, err)
// 		return ErrCreateUser
// 	}

// 	Logr.Infof("Created User %s", name)
// 	return nil
// }

// InitUser initializes the client to be used by the main program
// and call the SmartContracts. `name` represents the identity of the user.
// It may present a `secret` for the enrollment.
// It should have been registered by the consortium previously.
func (fs *FabricSetup) InitUser(name string, secret ...string) error {
	Logr.Debugf("FabricSet.InitUser entered for %s", name)
	if name == fs.currentUser {
		// already the right client
		return nil
	}

	var err error
	if len(secret) == 0 {
		// err = fs.initUser(name)
	} else {
		err = fs.initUserWithSecret(name, secret[0])
	}
	if err != nil {
		return ErrInitClient
	}

	// Channel client is used to query and execute transactions
	clientContext := fs.sdk.ChannelContext(fs.ChannelID, fabsdk.WithUser(name), fabsdk.WithOrg(fs.OrgName))
	fs.client, err = channel.New(clientContext)
	if err != nil {
		Logr.Debugf("channelID %s name %s org %s", fs.ChannelID, name, fs.OrgName)
		Logr.Errorf("failed to create new channel client for user %s  %v", name, err)
		return ErrInitClient
	}

	// Creation of the client which will enables access to our channel events
	fs.event, err = event.New(clientContext)
	if err != nil {
		Logr.Errorf("failed to create new event client for user %s %v", name, err)
		return ErrInitClient
	}

	// Everyting is OK.
	fs.currentUser = name
	Logr.Debugf("FabricSet.InitUser succeded for %s", fs.currentUser)
	return nil
}

// -------------------------------------
//
// ------------------------------

// getPrivateKeyName retruns the name of the private key of user `name` in the key store.
func (fs *FabricSetup) getPrivateKeyName(name string) (string, error) {

	ctxProvider2 := fs.sdk.Context(fabsdk.WithOrg(fs.OrgName))
	mspClient2, err := msp.New(ctxProvider2)
	if err != nil {
		Logr.Fatalf("getPrivateKeyName: Failed to init client: %v", err)
		return "", ErrInitUser
	}

	si, err := mspClient2.GetSigningIdentity(name)
	if err != nil {
		Logr.Fatalf("getPrivateKeyName: could not get signing identity of %s: %v", name, err)
		return "", ErrInitUser
	}

	a := si.PrivateKey().SKI()
	s := hex.EncodeToString(a) + "_sk"
	return s, nil
}

// // initUser initializes the user.  If the credentials are not present, it attempts
// // to reenroll the user, thus getting the key pair.
// func (fs *FabricSetup) initUser(name string) error {

// 	ctxProvider2 := fs.sdk.Context(fabsdk.WithOrg(fs.OrgName))
// 	mspClient2, err := msp.New(ctxProvider2)
// 	if err != nil {
// 		Logr.Fatalf("initUser: Failed to init client: %v", err)
// 		return ErrInitUser
// 	}

// 	// checks whether the user is not already known
// 	_, err = mspClient2.GetSigningIdentity(name)
// 	if err == nil {
// 		Logr.Infof("%s already exist.  Skip the init.", name)
// 		return nil
// 	}

// 	// retrieves the user
// 	secret, err := getSecret(name)
// 	Logr.Debugf("initUser %s with %s from vault %s", name, secret, store)
// 	if err != nil {
// 		Logr.Errorf("%s is not known by the vault %s.", name, store)
// 		return errors.Wrapf(ErrInitUser, "%s is not known by the vault.", name)
// 	}

// 	err = mspClient2.Enroll(name, msp.WithSecret(secret))
// 	if err != nil {
// 		Logr.Errorf("Could not reenroll %s due to %v", name, err)
// 		return ErrInitUser
// 	}

// 	Logr.Infof("Enrolled user %s", name)
// 	return nil
// }

// initUserWithSecret initializes the user `name` using the enrolment secret
// `secret`.  If the credentials are not present, it attempts
// to reenroll the user, thus getting the key pair.
func (fs *FabricSetup) initUserWithSecret(name string, secret string) error {

	ctxProvider2 := fs.sdk.Context(fabsdk.WithOrg(fs.OrgName))
	mspClient2, err := msp.New(ctxProvider2)
	if err != nil {
		Logr.Fatalf("initUser: Failed to init client: %v", err)
		return ErrInitUser
	}

	// checks whether the user is not already known
	_, err = mspClient2.GetSigningIdentity(name)
	if err == nil {
		Logr.Infof("%s already exist.  Skip the init.", name)
		return nil
	}

	err = mspClient2.Enroll(name, msp.WithSecret(secret))
	if err != nil {
		Logr.Errorf("Could not reenroll %s due to %v", name, err)
		return ErrInitUser
	}

	Logr.Infof("Enrolled user %s", name)
	return nil
}

// // ------------------------
// // Private functions
// // --------------------
// // ----

// func addAttributes(attri []msp.Attribute, name string, value string) []msp.Attribute {
// 	var attr msp.Attribute

// 	attr.Name = name
// 	attr.Value = value
// 	attr.ECert = true

// 	return append(attri, attr)
// }

// func generateReq(name, roleName string, attr ...Attribute) *msp.RegistrationRequest {

// 	req := &msp.RegistrationRequest{}
// 	req.Name = name
// 	req.Type = "client" // to be aceptable with OU, it should be client (not user as said in doc)

// 	req.Attributes = addAttributes(req.Attributes, "Cert", name)
// 	req.Attributes = addAttributes(req.Attributes, "Role", roleName)

// 	for _, a := range attr {
// 		if a.Key == "Cert" || a.Key == "Role" {
// 			Logr.Infof("CreateUSer: rejected attribute %s:%s as redundant",
// 				a.Key, a.Value)
// 		} else {
// 			req.Attributes = addAttributes(req.Attributes, a.Key, a.Value)
// 		}
// 	}
// 	return req
// }
