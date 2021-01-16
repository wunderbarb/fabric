// V0.3.5
// Author: DIEHL E.
// (C) Sony Pictures Entertainment, Nov 2020

package blockchain

import (
	"errors"
)

var (
	// ErrClientNotInitialized occurs when the Client is invoked before being
	// initialized.  This may happen if using a structure Client that was not created
	// by NewClient.
	ErrClientNotInitialized = errors.New("client is not initialized")
	// ErrCreateUser occurs when the blockchain could not invoke
	// the creation of a user.
	ErrCreateUser = errors.New("cannot create new user")
	// ErrNotProperNetwork occurs when the toml file does not point to a
	// supported cloud.
	ErrNotProperNetwork = errors.New("wrong network in toml file ")
	// ErrPathChaincode occurs when ChaincodePath and ChaincodePackage are either
	// both defined or both null.
	ErrPathChaincode = errors.New("could not decide between ChaincodePath and ChaincodePackage")
	// ErrSDKInitialized occurs trying to reinitialize the SDK.
	ErrSDKInitialized = errors.New("sdk already initialized ")
	// ErrSDKFailed occurs when trying to initialize fabric-sdk-go SDK.  More information in the log.
	ErrSDKFailed = errors.New("sdk failed initialization ")
	// ErrInitClient occurs when trying to initaite a SDK client. More information in the log.
	ErrInitClient = errors.New("sdk failed to init the client")
	// ErrInitUser occurs when trying to initaite a SDK user. More information in the log.
	ErrInitUser = errors.New("sdk failed to init the user client")
	// ErrFailedChannelInit occurs when trying to instantiate and install a channel.
	// More information in the log.
	ErrFailedChannelInit = errors.New("failed to instantiate the channel")
	// ErrFailedChaincodeInstall occurs when attempting to install a chaincode on a channel
	ErrFailedChaincodeInstall = errors.New("failed to install the chaincode")
	// ErrNoVault occurs when attempting to open a vault file that does not
	// exist.
	ErrNoVault = errors.New("no vault file")
	// ErrUpgradeButNoConfig occurs when upgrade was OK but failed to overwrite the config file,
	// when using option inViper in UpgradeCC.
	ErrUpgradeButNoConfig = errors.New("could not update configfile")
	// ErrWalletInitFailed occurs when the wallet cannot be started or populated.
	ErrWalletInitFailed = errors.New("could not init the wallet")
	// ErrUserAlreadyExist occurs when the user is already known and it is required
	// to be created again.
	ErrUserAlreadyExist = errors.New("user already exists")
)
