package main

import (
	"fmt"
	"log"
	"os"
)

const (
	EAX = iota
	ECX
	EDX
	EBX
	ESP
	EBP
	ESI
	EDI
	registersCount
	AL = EAX
	CL = ECX
	DL = EDX
	BL = EBX
	AH = AL + 4
	CH = CL + 4
	DH = DL + 4
	BH = BL + 4
)

const (
	CarryFlag    uint32 = 1
	ZeroFlag     uint32 = 1 << 6
	SignFlag     uint32 = 1 << 7
	OverflowFlag uint32 = 1 << 11
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

func (emu *Emulator) getRm8(modrm *ModRM) uint8 {
	if modrm.mod == 3 {
		return emu.getRegister8(modrm.rm)
	} else {
		address := emu.calcMemoryAddress(modrm)
		return emu.getMemory8(address)
	}
}

func (emu *Emulator) getRm32(modrm *ModRM) uint32 {
	if modrm.mod == 3 {
		return emu.getRegister32(modrm.rm)
	} else {
		address := emu.calcMemoryAddress(modrm)
		return emu.getMemory32(address)
	}
}

func (emu *Emulator) setRm8(modrm *ModRM, val uint8) {
	if modrm.mod == 3 {
		emu.setRegister8(modrm.rm, val)
	} else {
		address := emu.calcMemoryAddress(modrm)
		emu.setMemory8(address, val)
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

func (emu *Emulator) getRegister8(index uint8) uint8 {
	if index < 4 {
		return uint8(emu.registers[index])
	} else {
		return uint8(emu.registers[index-4] >> 8)
	}
}

func (emu *Emulator) setRegister8(index uint8, val uint8) {
	if index < 4 {
		emu.registers[index] = emu.registers[index]&0xffffff00 | uint32(val)
	} else {
		emu.registers[index] = emu.registers[index-4]&0xffff00ff | uint32(val)<<8
	}
}

func (emu *Emulator) getRegister32(index uint8) uint32 {
	return emu.registers[index]
}

func (emu *Emulator) setRegister32(index uint8, val uint32) {
	emu.registers[index] = val
}

func (emu *Emulator) push32(val uint32) {
	address := emu.getRegister32(ESP) - 4
	emu.setRegister32(ESP, address)
	emu.setMemory32(address, val)
}

func (emu *Emulator) pop32() uint32 {
	address := emu.getRegister32(ESP)
	mem := emu.getMemory32(address)
	emu.setRegister32(ESP, address+4)
	return mem
}

func (emu *Emulator) updateEflagsSub(v1 uint32, v2 uint32) {
	sign1 := v1 >> 31
	sign2 := v2 >> 31
	res := uint64(v1) - uint64(v2)
	signr := uint32(res>>31) & 1

	emu.setCarry(res>>32 != 0)
	emu.setZero(res == 0)
	emu.setSign(signr == 1)
	emu.setOverflow(sign1 != sign2 && sign1 != signr)
}

func (emu *Emulator) setCarry(isCarry bool) {
	if isCarry {
		emu.eflags |= CarryFlag
	} else {
		emu.eflags &= ^CarryFlag
	}
}

func (emu *Emulator) setZero(isZero bool) {
	if isZero {
		emu.eflags |= ZeroFlag
	} else {
		emu.eflags &= ^ZeroFlag
	}
}

func (emu *Emulator) setSign(isSign bool) {
	if isSign {
		emu.eflags |= SignFlag
	} else {
		emu.eflags &= ^SignFlag
	}
}

func (emu *Emulator) setOverflow(isOverflow bool) {
	if isOverflow {
		emu.eflags |= OverflowFlag
	} else {
		emu.eflags &= ^OverflowFlag
	}
}

func (emu *Emulator) isCarry() bool {
	return emu.eflags&CarryFlag != 0
}

func (emu *Emulator) isZero() bool {
	return emu.eflags&ZeroFlag != 0
}

func (emu *Emulator) isSign() bool {
	return emu.eflags&SignFlag != 0
}

func (emu *Emulator) isOverflow() bool {
	return emu.eflags&OverflowFlag != 0
}
