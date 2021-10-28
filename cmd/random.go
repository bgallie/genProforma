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
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/bgallie/tntengine"
	"github.com/spf13/cobra"
)

// randomCmd represents the random command
var randomCmd = &cobra.Command{
	Use:   "random",
	Short: "Generate a new random proforma machine",
	Long:  `Generate a new random proforma machine using Go's cryptographically secure random number generator.`,
	Run: func(cmd *cobra.Command, args []string) {
		generatRandomMachine()
	},
}

func init() {
	rootCmd.AddCommand(randomCmd)
	cycleSizes = perm(len(tntengine.CycleSizes))
	rotorSizesIndex = perm(len(rotorSizes))
}

func perm(n int) []int {
	res := make([]int, n)

	for i := range res {
		res[i] = i
	}

	for i := (n - 1); i > 0; i-- {
		rnd, err := rand.Int(rand.Reader, big.NewInt(int64(i)))
		if err != nil {
			log.Fatalln(err)
		}
		j := int(rnd.Int64())
		res[i], res[j] = res[j], res[i]
	}

	return res
}

func randP() []byte {
	res := make([]byte, 256)

	for i := range res {
		res[i] = byte(i)
	}

	for i := (256 - 1); i > 0; i-- {
		rnd, err := rand.Int(rand.Reader, big.NewInt(int64(i)))
		if err != nil {
			log.Fatalln(err)
		}
		j := int(rnd.Int64())
		res[i], res[j] = res[j], res[i]
	}

	return res
}

func updateRotor(r *tntengine.Rotor) {
	r.Size = rotorSizes[rotorSizesIndex[rCnt]]
	rnd, err := rand.Int(rand.Reader, big.NewInt(int64(r.Size)))
	if err != nil {
		log.Fatalln(err)
	}
	r.Start = int(rnd.Int64())
	r.Current = r.Start
	rnd, err = rand.Int(rand.Reader, big.NewInt(int64(r.Size)))
	if err != nil {
		log.Fatalln(err)
	}
	r.Step = int(rnd.Int64())
	// blkCnt is the total number of bytes needed to hold rotorSize bits + a slice of 256 bits
	blkCnt := ((r.Size + tntengine.CypherBlockSize + 7) / 8)
	r.Rotor = make([]byte, blkCnt)
	_, err = rand.Read(r.Rotor)
	if err != nil {
		log.Fatalln(err)
	}

	//Slice the first 256 bits of the rotor to the end of the rotor
	var j = r.Size
	for i := 0; i < 256; i++ {
		if tntengine.GetBit(r.Rotor, uint(i)) {
			tntengine.SetBit(r.Rotor, uint(j))
		} else {
			tntengine.ClrBit(r.Rotor, uint(j))
		}
		j++
	}

	rCnt++
}

func updatePermutator(p *tntengine.Permutator) {
	p.Randp = randP()
	p.Cycles = make([]tntengine.Cycle, tntengine.NumberPermutationCycles)

	for i := range p.Cycles {
		p.Cycles[i].Length = tntengine.CycleSizes[cycleSizes[pCnt]][i]
		p.Cycles[i].Current = 0
		// Adjust the start to reflect the lenght of the previous cycles
		if i == 0 { // no previous cycle so start at 0
			p.Cycles[i].Start = 0
		} else {
			p.Cycles[i].Start = p.Cycles[i-1].Start + p.Cycles[i-1].Length
		}
	}

	p.CurrentState = 0
	p.MaximalStates = p.Cycles[0].Length

	for i := 1; i < len(p.Cycles); i++ {
		p.MaximalStates *= p.Cycles[i].Length
	}

	pCnt++
}

func generatRandomMachine() {
	var err error
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
			if err != nil {
				log.Fatalln(err)
			}
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
			if err != nil {
				log.Fatalln(err)
			}
		case *tntengine.Permutator:
			err = jEncoder.Encode(machine.(*tntengine.Permutator))
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}
