// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package util

import (
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/iotex-antenna-go/v2/account"
)

// MustFetchNonEmptyParam must fetch a nonempty environment variable
func MustFetchNonEmptyParam(key string) string {
	str := os.Getenv(key)
	if len(str) == 0 {
		log.Fatalf("%s is not defined in env\n", key)
	}
	return str
}

// GetVaultAccount returns the vault account given the password
func GetVaultAccount(pwd string) (account.Account, error) {
	// load the keystore file
	ks := keystore.NewKeyStore("./", keystore.StandardScryptN, keystore.StandardScryptP)
	if len(ks.Accounts()) != 1 {
		return nil, fmt.Errorf("found %d keys, expecting 1", len(ks.Accounts()))
	}
	pk, err := crypto.KeystoreToPrivateKey(ks.Accounts()[0], pwd)
	if err != nil {
		return nil, fmt.Errorf("error decrypting the vault private key")
	}
	return account.PrivateKeyToAccount(pk)
}
