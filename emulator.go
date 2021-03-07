package main

import (
	"fmt"
	"log"
	"os"
)

type Register uint8

const (
	EAX Register = iota
	ECX
	EDX
	EBX
	ESP
	EBP
	ESI
	EDI
	registersCount
)

var (
	registersName []string = []string{"EAX", "ECX", "EDX", "EBX", "ESP", "EBP", "ESI", "EDI"}
)

type Emulator struct {
	registers [registersCount]uint32
	eflags    uint32
	memory    []uint8
	eip       uint32
}

func NewEmulator(size uint, eip uint32, esp uint32) *Emulator {
	emu := new(Emulator)
	emu.memory = make([]uint8, size)
	emu.eip = eip
	emu.registers[ESP] = esp
	return emu
}

func (emu *Emulator) getCode8(index int) uint8 {
	return emu.memory[int(emu.eip)+index]
}

func (emu *Emulator) getSignCode8(index int) int8 {
	return int8(emu.memory[int(emu.eip)+index])
}

func (emu *Emulator) getCode32(index int) uint32 {
	var ret uint32
	for i := 0; i < 4; i++ {
		ret |= uint32(emu.getCode8(index+i)) << (i * 8)
	}
	return ret
}

func (emu *Emulator) getSignCode32(index int) int32 {
	return int32(emu.getCode32(index))
}

func (emu *Emulator) dumpRegisters() {
	for i := 0; i < int(registersCount); i++ {
		fmt.Printf("%s = %08x\n", registersName[i], emu.registers[i])
	}
	fmt.Printf("EIP = %08x\n", emu.eip)
}

func (emu *Emulator) getRm32(modrm *ModRM) uint32 {
	if modrm.mod == 3 {
		return emu.getRegister32(modrm.rm)
	} else {
		address := emu.calcMemoryAddress(modrm)
		return emu.getMemory32(address)
	}
}

func (emu *Emulator) setRm32(modrm *ModRM, val uint32) {
	if modrm.mod == 3 {
		emu.setRegister32(modrm.rm, val)
	} else {
		address := emu.calcMemoryAddress(modrm)
		emu.setMemory32(address, val)
	}
}

func (emu *Emulator) getMemory8(address uint32) uint8 {
	return emu.memory[address]
}

func (emu *Emulator) getMemory32(address uint32) uint32 {
	var mem uint32
	for i := uint32(0); i < 4; i++ {
		mem |= uint32(emu.getMemory8(address+i)) << (i * 8)
	}
	return mem
}

func (emu *Emulator) setMemory8(address uint32, val uint8) {
	emu.memory[address] = val
}

func (emu *Emulator) setMemory32(address uint32, val uint32) {
	for i := uint32(0); i < 4; i++ {
		emu.setMemory8(address+i, uint8(val>>(i*8)))
	}
}

func (emu *Emulator) calcMemoryAddress(modrm *ModRM) uint32 {
	if modrm.rm == 4 {
		log.Println("not implemented ModRM rm = 4")
		os.Exit(0)
	}

	switch modrm.mod {
	case 0:
		if modrm.rm == 5 {
			return modrm.disp32
		}
		return emu.getRegister32(modrm.rm)
	case 1:
		return emu.getRegister32(modrm.rm) + uint32(modrm.disp8)
	case 2:
		return emu.getRegister32(modrm.rm) + modrm.disp32
	case 3:
		log.Println("not implemented ModRM mod = 3")
		os.Exit(0)
	}
	return 0
}

func (emu *Emulator) getRegister32(index uint8) uint32 {
	return emu.registers[index]
}

func (emu *Emulator) setRegister32(index uint8, val uint32) {
	emu.registers[index] = val
}

func (emu *Emulator) push32(val uint32) {
	address := emu.getRegister32(uint8(ESP)) - 4
	fmt.Printf("%x %x\n", address, val)
	emu.setRegister32(uint8(ESP), address)
	emu.setMemory32(address, val)
}

func (emu *Emulator) pop32() uint32 {
	address := emu.getRegister32(uint8(ESP))
	mem := emu.getMemory32(address)
	fmt.Printf("%x %x\n", address, mem)
	emu.setRegister32(uint8(ESP), address+4)
	return mem
}
