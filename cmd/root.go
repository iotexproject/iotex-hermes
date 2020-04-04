// Copyright (c) 2020 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/iotexproject/iotex-hermes/cmd/claim"
	"github.com/iotexproject/iotex-hermes/cmd/distribute"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "hermes",
	Short: "Command-line interface for IoTeX rewards distribution tool",
	Long:  "Command-line interface for IoTeX rewards distribution tool",
}

func init() {
	RootCmd.AddCommand(claim.ClaimCmd)
	RootCmd.AddCommand(distribute.DistributeCmd)
}
