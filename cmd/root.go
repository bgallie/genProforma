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
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/bgallie/tntengine"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	cfgFile string
	// rotoSizes is an array of possible rotor sizes.  It consists of prime
	// numbers less than 1792 to allow for a 256 bit splce at the end of the
	// rotor and still be less then or equal to 2048 bits (32 bytes).  The rotor
	// sizes selected from this list will maximizes the number of unique states
	// the rotors can take.
	rotorSizes = []int{
		1669, 1693, 1697, 1699, 1709, 1721, 1723, 1733,
		1741, 1747, 1753, 1759, 1777, 1783, 1787, 1789}
	rotorSizesIndex []int
	cycleSizes      []int
	outputFileName  string
	rotor1          = new(tntengine.Rotor)
	rotor2          = new(tntengine.Rotor)
	rotor3          = new(tntengine.Rotor)
	rotor4          = new(tntengine.Rotor)
	rotor5          = new(tntengine.Rotor)
	rotor6          = new(tntengine.Rotor)
	permutator1     = new(tntengine.Permutator)
	permutator2     = new(tntengine.Permutator)
	proFormaMachine = []tntengine.Crypter{rotor1, rotor2, permutator1, rotor3, rotor4, permutator2, rotor5, rotor6}
	rCnt            = 0
	pCnt            = 0
	outputFile      *os.File
	GitCommit       string = "not set"
	GitBranch       string = "not set"
	GitState        string = "not set"
	GitSummary      string = "not set"
	BuildDate       string = "not set"
	Version         string = "dev"
	rRead           func([]byte) (n int, err error)
	rPerm           func(int) []int
	rInt            func(int64) int64
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "genProforma",
	Short: "Generate proforma rotors and permutators.",
	Long: `genProfroma is a tool to generates a set of rotors and permutators that 
	can be used by tntengine to override the builtin proforma rotors
	and permutators.`,
	Version: Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.genProforma.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFileName, "outputfile", "f", "-", "output file to write the proforma rotors and permutators to")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".genProforma" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".genProforma")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func perm(n int) []int {
	res := make([]int, n)

	for i := range res {
		res[i] = i
	}

	for i := (n - 1); i > 0; i-- {
		j := rInt(int64(i))
		res[i], res[j] = res[j], res[i]
	}

	return res
}

func randP() []byte {
	res := make([]byte, tntengine.CypherBlockSize)

	// Create a table of byte values [0...255] in a random order
	for i, val := range rPerm(tntengine.CypherBlockSize) {
		res[i] = byte(val)
	}

	return res
}

func updateRotor(r *tntengine.Rotor) {
	size := rotorSizes[rotorSizesIndex[rCnt]]
	start := int(rInt(int64(size)))
	step := int(rInt(int64(size-1))) + 1
	// blkCnt is the total number of bytes needed to hold rotorSize bits + a slice of 256 bits
	blkCnt := ((size + tntengine.CypherBlockSize + 7) / 8)
	rotor := make([]byte, blkCnt)
	_, err := rRead(rotor)
	cobra.CheckErr(err)
	r.New(size, start, step, rotor)
	rCnt++
}

func updatePermutator(p *tntengine.Permutator) {
	p.New(tntengine.CycleSizes[cycleSizes[pCnt]][:], randP())
	pCnt++
}

func generatRandomMachine() {
	var err error
	cycleSizes = perm(len(tntengine.CycleSizes))
	rotorSizesIndex = perm(len(rotorSizes))
	// Update the rotors and permutators in a very non-linear fashion.
	for _, machine := range proFormaMachine {
		switch v := machine.(type) {
		default:
			fmt.Fprintf(os.Stderr, "Unknown machine: %v\n", v)
		case *tntengine.Rotor:
			updateRotor(machine.(*tntengine.Rotor))
		case *tntengine.Permutator:
			updatePermutator(machine.(*tntengine.Permutator))
		case *tntengine.Counter:
			machine.(*tntengine.Counter).SetIndex(big.NewInt(0))
		}
	}

	if len(outputFileName) != 0 {
		if outputFileName == "-" {
			outputFile = os.Stdout
		} else {
			outputFile, err = os.Create(outputFileName)
			cobra.CheckErr(err)
		}
	}

	defer outputFile.Close()
	jEncoder := json.NewEncoder(outputFile)
	jEncoder.SetEscapeHTML(false)

	for _, machine := range proFormaMachine {
		switch v := machine.(type) {
		default:
			fmt.Fprintf(os.Stderr, "Unknown machine: %v\n", v)
		case *tntengine.Rotor:
			err = jEncoder.Encode(machine.(*tntengine.Rotor))
			cobra.CheckErr(err)
		case *tntengine.Permutator:
			err = jEncoder.Encode(machine.(*tntengine.Permutator))
			cobra.CheckErr(err)
		}
	}
}
