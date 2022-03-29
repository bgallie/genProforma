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

	"github.com/bgallie/tntengine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	tntMachine tntengine.TntEngine
	random     *tntengine.Rand
)

// tntengineCmd represents the tntengine command
var tntengineCmd = &cobra.Command{
	Use:   "tntengine",
	Short: "Generate a new proforma machine",
	Long:  `Generate a new proforma machine using a psudo-random number generator (tntengine).`,
	Run: func(cmd *cobra.Command, args []string) {
		initEngine(args)
		generatRandomMachine()
	},
}

func init() {
	rootCmd.AddCommand(tntengineCmd)
}

func initEngine(args []string) {
	// Obtain the passphrase used to encrypt the file from either:
	// 1. User input from the terminal (most secure)
	// 2. The 'TNT2_SECRET' environment variable (less secure)
	// 3. Arguments from the entered command line (least secure - not recommended)
	var secret string
	if len(args) == 0 {
		if viper.IsSet("TNT2_SECRET") {
			secret = viper.GetString("TNT2_SECRET")
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

	// Initialize the tntengine with the secret key and the named proforma file.
	tntMachine.Init([]byte(secret), "")
	tntMachine.SetEngineType("E")
	// Now the the engine type is set, build the cipher machine.
	tntMachine.BuildCipherMachine()
	// Get the random functions
	random = new(tntengine.Rand).New(&tntMachine)
	rRead = random.Read
	rPerm = random.Perm
	rInt = random.Int63n
}
