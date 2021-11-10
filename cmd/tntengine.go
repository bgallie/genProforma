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
	rRead      func([]byte) (n int, err error)
	rInt       func(int64) int64
)

// tntengineCmd represents the tntengine command
var tntengineCmd = &cobra.Command{
	Use:   "tntengine",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		initEngine(args)
		cycleSizes = perm(len(tntengine.CycleSizes))
		rotorSizesIndex = perm(len(rotorSizes))
		generatRandomMachine()
	},
}

func init() {
	rootCmd.AddCommand(tntengineCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tntengineCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tntengineCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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
	random = tntengine.NewRand(&tntMachine)
	rRead = random.Read
	rInt = random.Int63n
}
