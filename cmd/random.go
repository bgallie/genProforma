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
	"crypto/rand"
	"math/big"

	"github.com/spf13/cobra"
)

// randomCmd represents the random command
var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Generate a new proforma machine",
	Long:  `Generate a new proforma machine using Go's cryptographically secure random number generator.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !rootCmd.Flags().Changed("outputType") {
			rootCmd.Flags().Set("outputType", "json")
		}
		generateProForma(outputType)
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
	// Define rRead and rInt to use the crypto/rand based functions.
	rRead = rand.Read
	rPerm = perm
	rInt = cInt
}

func cInt(n int64) int64 {
	max := big.NewInt(n)
	j, err := rand.Int(rand.Reader, max)
	cobra.CheckErr(err)
	return j.Int64()
}
