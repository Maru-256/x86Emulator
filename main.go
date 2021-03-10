package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	memorySize = 1 << 20
)

func main() {
	var quiet bool
	flag.BoolVar(&quiet, "q", false, "do not output current EIP and opcode")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalln("usage: px86 [-q] filename")
	}

	const org = 0x7C00
	emu := NewEmulator(memorySize, org, org)
	b, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}
	if err := copyByIndex(emu.memory, b, emu.eip); err != nil {
		log.Fatalln(err)
	}

	for emu.eip < memorySize {
		code := emu.getCode8(0)
		if !quiet {
			fmt.Printf("EIP = %X, Code = %02X\n", emu.eip, code)
		}

		if instructions[code] == nil {
			log.Printf("Not Implemented: %x\n", code)
			break
		}
		instructions[code](emu)
		if emu.eip == 0x00 {
			log.Println("end of program")
			break
		}
	}

	emu.dumpRegisters()
}

func copyByIndex(dst, src []uint8, index uint32) error {
	for i := uint32(0); i < uint32(len(src)); i++ {
		if index+uint32(len(src)) > uint32(len(dst)) {
			return fmt.Errorf("out of index")
		}
		dst[index+i] = src[i]
	}

	return nil
}
