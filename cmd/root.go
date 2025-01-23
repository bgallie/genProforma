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
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "genProforma",
	Short: "Generate proforma rotors and permutators.",
	Long: `genProfroma is a tool to generates a set of rotors and permutators that 
	can be used to override the builtin proforma rotors
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
	rootCmd.PersistentFlags().StringVarP(&outputType, "outputType", "t", "-", `Output type to generate.
	The valid types are "json" (default) and "ikm" (default for ikmachine command).
	    json: outputs JSON encoded format.
	    ikm: outputs a string in valid golang that can replace the proforma rotors and permutators in ikmachine/machine.go`)
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

var (
	cfgFile string
	// rotoSizes is an array of possible rotor sizes.  It consists of prime
	// numbers less than 1792 to allow for a 256 bit splce at the end of the
	// rotor and still be less then or equal to 2048 bits (32 bytes).  The rotor
	// sizes selected from this list will maximizes the number of unique states
	// the rotors can take.
	rotorSizes      = []int16{1789, 1787, 1777, 1759, 1753, 1747}
	cycleSizes      = CycleSizes{61, 63, 65, 67}
	outputFileName  string
	rotors          = []*Rotor{new(Rotor), new(Rotor), new(Rotor), new(Rotor), new(Rotor), new(Rotor)}
	permutators     = []*Permutator{new(Permutator), new(Permutator)}
	proformaMachine = make([]any, len(rotors)+len(permutators))
	rCnt            = 0
	pCnt            = 0
	outputFile      *os.File
	rRead           func([]byte) (n int, err error)
	rPerm           func(int) []int
	rInt            func(int64) int64
	outputType      string
	prefix          string = ""
)

// CycleSizes contains the cycle sizes used by the permutators.
type CycleSizes []int16

// Rotor is the type of a rotor used in IkMachine
type Rotor struct {
	Size    int16  // the size in bits for this rotor
	Start   int16  // the initial starting position of the rotor
	Step    int16  // the step size in bits for this rotor
	Current int16  // the current position of this rotor
	Rotor   []byte // the rotor
}

// String converts a Rotor to a string representation of the Rotor.
func (r *Rotor) String() string {
	var output bytes.Buffer
	rotorLen := len(r.Rotor)
	output.WriteString(prefix + "{\n")
	output.WriteString(fmt.Sprintf("%s\tsize:    %d,\n", prefix, r.Size))
	output.WriteString(fmt.Sprintf("%s\tstart:   %d,\n", prefix, r.Start))
	output.WriteString(fmt.Sprintf("%s\tstep:    %d,\n", prefix, r.Step))
	output.WriteString(fmt.Sprintf("%s\tcurrent: %d,\n", prefix, r.Current))
	output.WriteString(prefix + "\trotor:   []byte{\n")
	for i := 0; i < rotorLen; i += 16 {
		output.WriteString(prefix + "\t\t")
		if i+16 < rotorLen {
			for _, k := range r.Rotor[i : i+15] {
				output.WriteString(fmt.Sprintf("%#02x, ", k))
			}
			output.WriteString(fmt.Sprintf("%#02x,\n", r.Rotor[i+15]))
		} else {
			l := len(r.Rotor[i:])
			for _, k := range r.Rotor[i : i+l-1] {
				output.WriteString(fmt.Sprintf("%#02x, ", k))
			}
			output.WriteString(fmt.Sprintf("%#02x}}", r.Rotor[i+l-1]))
		}
	}
	return output.String()
}

// Cycle describes a cycle for the permutator so it can adjust the permutation
// table used to permute the block.  IkMachine currently uses a single cycle to
// rearrange Randp into bitPerm
type Cycle struct {
	Start   int16 // The starting point (into randp) for this cycle.
	Length  int16 // The length of the cycle.
	Current int16 // The point in the cycle [0 .. cycles.length-1] to start
}

// Permutator is a type that defines a permutation used in IkMachine.
type Permutator struct {
	CurrentState  int32     // Current number of cycles for this permutator.
	MaximalStates int32     // Maximum number of cycles this permutator can have before repeating.
	Cycles        []Cycle   // Cycles ordered by the current permutation.
	Randp         []byte    // Values 0 - 255 in a random order.
	bitPerm       [256]byte // Permutation table created from Randp.
}

// String formats a string representing the permutator (as Go source code).
func (p *Permutator) String() string {
	var output bytes.Buffer
	output.WriteString(prefix + "{\n")
	output.WriteString(fmt.Sprintf(prefix+"\tcurrentState:  %d,\n", p.CurrentState))
	output.WriteString(fmt.Sprintf(prefix+"\tmaximalStates: %d,\n", p.MaximalStates))
	output.WriteString(prefix + "\tcycles: []Cycle{\n")
	var i int
	if len(p.Cycles) > 1 {
		for i = range len(p.Cycles) - 1 {
			output.WriteString(prefix + "\t\t")
			output.WriteString(fmt.Sprintf("{start: %d, length: %d, current: %d},\n",
				p.Cycles[i].Start, p.Cycles[i].Length, p.Cycles[i].Current))
		}
		i++ // make i the index of the last permutation cycle.
	}
	output.WriteString(prefix + "\t\t")
	output.WriteString(fmt.Sprintf("{start: %d, length: %d, current: %d},\n",
		p.Cycles[i].Start, p.Cycles[i].Length, p.Cycles[i].Current))
	output.WriteString(prefix + "\t},\n" + prefix + "\trandp: []byte{\n")
	for i := 0; i < 256; i += 16 {
		output.WriteString(prefix + "\t\t")
		if i != (256 - 16) {
			for _, k := range p.Randp[i : i+15] {
				output.WriteString(fmt.Sprintf("%#02x, ", k))
			}
			output.WriteString(fmt.Sprintf("%#02x,\n", p.Randp[i+15]))
		} else {
			for _, k := range p.Randp[i : i+15] {
				output.WriteString(fmt.Sprintf("%#02x, ", k))
			}
			output.WriteString(fmt.Sprintf("%#02x},\n", p.Randp[i+15]))
		}
	}
	output.WriteString(prefix + "\tbitPerm: [256]byte{\n")
	for i := 0; i < 256; i += 16 {
		output.WriteString(prefix + "\t\t")
		if i != (256 - 16) {
			for _, k := range p.bitPerm[i : i+15] {
				output.WriteString(fmt.Sprintf("%#02x, ", k))
			}
			output.WriteString(fmt.Sprintf("%#02x,\n", p.bitPerm[i+15]))
		} else {
			for _, k := range p.bitPerm[i : i+15] {
				output.WriteString(fmt.Sprintf("%#02x, ", k))
			}
			output.WriteString(fmt.Sprintf("%#02x}}", p.bitPerm[i+15]))
		}
	}
	return output.String()
}

func perm(n int) []int {
	if n < 0 {
		panic(fmt.Sprintf("Perm called with a negative argument [%d]", n))
	}
	res := make([]int, n)
	for i := 1; i < n; i++ {
		j := rInt(int64(i + 1))
		res[i] = res[j]
		res[j] = i
	}
	return res
}

func randP() []byte {
	res := make([]byte, 256)

	// Create a table of byte values [0...255] in a random order
	for i, val := range rPerm(256) {
		res[i] = byte(val)
	}

	return res
}

func updateRotor(r *Rotor) {
	r.Size = rotorSizes[rCnt]
	r.Start = int16(rInt(int64(r.Size)))
	r.Step = int16(rInt(int64(r.Size-1))) + 1
	// blkCnt is the total number of bytes needed to hold rotorSize bits + a slice of 256 bits
	blkCnt := ((r.Size + 7) / 8)
	r.Rotor = make([]byte, 256)
	rData := make([]byte, blkCnt)
	_, err := rRead(rData)
	cobra.CheckErr(err)
	copy(r.Rotor, rData)
	sliceRotor(r)
	rCnt++
}

func updatePermutator(p *Permutator) {
	cycleOrder := perm(len(cycleSizes))
	p.CurrentState = 0
	p.MaximalStates = 1
	p.Cycles = make([]Cycle, len(cycleSizes))
	runningLength := int16(0)
	for i := range p.Cycles {
		p.Cycles[i].Start = runningLength
		p.Cycles[i].Length = cycleSizes[cycleOrder[i]]
		p.MaximalStates *= int32(p.Cycles[i].Length)
		runningLength += p.Cycles[i].Length
	}
	p.Randp = randP()
	pCnt++
}

// sliceRotor appends the first 256 bits of the rotor to the end of the rotor.
func sliceRotor(r *Rotor) {
	var size, sBlk, sBit, Rshift, Lshift uint
	size = uint(r.Size)
	sBlk = size >> 3
	sBit = size & 7
	Rshift = 8 - sBit
	Lshift = sBit
	if sBit != 0 {
		// The copy appending will be done at the byte level instead of the bit level
		// so that we only loop 32 times instead of 256 times.
		for i := range 32 {
			r.Rotor[sBlk] &= (0xff >> Rshift)       // Clear out the bits that will be replaced
			r.Rotor[sBlk] |= (r.Rotor[i] << Lshift) // and add in the bits from the beginning of the rotor
			sBlk++
			r.Rotor[sBlk] = (r.Rotor[i] >> Rshift) // Seed the next byte at the end with the remaining bits from the beginning byte.
		}
	} else {
		copy(r.Rotor[sBlk:], r.Rotor[0:32])
	}
}

func generateProForma(oType string) {
	var err error
	updateRotor(rotors[0])
	proformaMachine[0] = rotors[0]
	updateRotor(rotors[1])
	proformaMachine[1] = rotors[1]
	updatePermutator(permutators[0])
	proformaMachine[2] = permutators[0]
	updateRotor(rotors[2])
	proformaMachine[3] = rotors[2]
	updateRotor(rotors[3])
	proformaMachine[4] = rotors[3]
	updatePermutator(permutators[1])
	proformaMachine[5] = permutators[1]
	updateRotor(rotors[4])
	proformaMachine[6] = rotors[4]
	updateRotor(rotors[5])
	proformaMachine[7] = rotors[5]
	if len(outputFileName) != 0 {
		if outputFileName == "-" {
			outputFile = os.Stdout
		} else {
			outputFile, err = os.Create(outputFileName)
			cobra.CheckErr(err)
		}
	}

	defer outputFile.Close()
	if oType == "json" {
		jEncoder := json.NewEncoder(outputFile)
		jEncoder.SetEscapeHTML(false)
		err = jEncoder.Encode(proformaMachine)
	} else {
		prefix = "\t\t"
		fmt.Fprint(outputFile,
			"\tproformaRotors = []*Rotor{\n\t\t// Define the proforma "+
				"rotors used to create the actual rotors to use.\n")
		for _, v := range rotors {
			fmt.Fprintf(outputFile, "%s,\n", v)
		}
		fmt.Fprint(outputFile, "\t}\n")
		fmt.Fprint(outputFile,
			"\tproformaPermutator = &Permutator{\n\t\t// Define the "+
				"proforma permutator used to create the actual permutator to use.\n")
		for _, v := range permutators {
			fmt.Fprintf(outputFile, "%s,\n", v)
		}
		fmt.Fprint(outputFile, "\t}\n")
	}
}
