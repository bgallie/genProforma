/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bgallie/ikmachine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	ikengine *ikmachine.IkMachine
	ikRandom *ikmachine.Rand
)

// tntengineCmd represents the tntengine command
var ikmachineCmd = &cobra.Command{
	Use:   "ikmachine",
	Short: "Generate a new proforma machine",
	Long:  `Generate a new proforma machine using a psudo-random number generator (ikmachine).`,
	Run: func(cmd *cobra.Command, args []string) {
		if rootCmd.Flags().Changed("outputType") {
			if outputType != "json" && outputType != "ikm" {
				cobra.CheckErr(outputType + " is not a valid output type.")
			}
		} else {
			rootCmd.Flags().Set("outputType", "ikm")
		}
		initEngine(args)
		generateProForma(outputType)
	},
}

func init() {
	rootCmd.AddCommand(ikmachineCmd)
}

func initIkEngine(args []string) {
	// Obtain the passphrase used to encrypt the file from either:
	// 1. User input from the terminal (most secure)
	// 2. The 'TNT2_SECRET' environment variable (less secure)
	// 3. Arguments from the entered command line (least secure - not recommended)
	var secret string
	if len(args) == 0 {
		if viper.IsSet("GPF_SECRET") {
			secret = viper.GetString("GPF_SECRET")
		} else {
			if term.IsTerminal(int(os.Stdin.Fd())) {
				fmt.Fprintf(os.Stderr, "Enter the passphrase: ")
				byteSecret, err := term.ReadPassword(int(os.Stdin.Fd()))
				cobra.CheckErr(err)
				fmt.Fprintln(os.Stderr, "")
				secret = string(byteSecret)
			}
		}
	} else {
		secret = strings.Join(args, " ")
	}

	if len(secret) == 0 {
		cobra.CheckErr("You must supply a password.")
		// } else {
		// 	fmt.Printf("Secret: [%s]\n", secret)
	}

	// Initialize the ikmachine with the secret key and the named proforma file.
	ikengine = new(ikmachine.IkMachine).InitializeProformaEngine().ApplyKey('E', []byte(secret))
	// Get the random functions
	ikRandom = new(ikmachine.Rand).New(ikengine)
	rRead = ikRandom.Read
	rPerm = ikRandom.Perm
	rInt = ikRandom.Int63n
}
