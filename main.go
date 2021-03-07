package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	memorySize = 1 << 20
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("usage: px86 filename")
	}

	emu := NewEmulator(memorySize, 0x7C00, 0x7C00)
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	if err := copyByIndex(emu.memory, b, 0x7C00); err != nil {
		log.Fatalln(err)
	}

	for emu.eip < memorySize {
		code := emu.getCode8(0)
		fmt.Printf("EIP = %X, Code = %02X\n", emu.eip, code)

		if instructions[code] == nil {
			log.Printf("Not Implemented: %x\n", code)
			break
		}
		instructions[code](emu)
		emu.dumpRegisters()
		if emu.eip == 0x00 {
			log.Println("end of program")
			break
		}
	}

	emu.dumpRegisters()
}

func copyByIndex(dst, src []uint8, index int) error {
	for i := 0; i < len(src); i++ {
		if index+len(src) > len(dst) {
			return fmt.Errorf("out of index")
		}
		dst[index+i] = src[i]
	}

	return nil
}
