package main

import (
	"log"
	"os"
)

var (
	instructions [256]instructionFunc
)

type instructionFunc func(emu *Emulator)

func init() {
	instructions[0x01] = addRm32R32

	instructions[0x3B] = cmpR32Rm32

	for i := 0; i < 8; i++ {
		instructions[0x50+i] = pushR32
		instructions[0x58+i] = popR32
	}

	instructions[0x68] = pushImm32
	instructions[0x6A] = pushImm8

	instructions[0x70] = jo
	instructions[0x71] = jno
	instructions[0x72] = jc
	instructions[0x73] = jnc
	instructions[0x74] = jz
	instructions[0x75] = jnz
	instructions[0x78] = js
	instructions[0x79] = jns
	instructions[0x7C] = jl
	instructions[0x7E] = jle

	instructions[0x83] = code83
	instructions[0x89] = movRm32R32
	instructions[0x8B] = movR32Rm32

	for i := 0; i < 8; i++ {
		instructions[0xB8+i] = movR32Imm32
	}

	instructions[0xC3] = ret
	instructions[0xC7] = movRm32Imm32
	instructions[0xC9] = leave

	instructions[0xE8] = callRel32
	instructions[0xE9] = nearJump
	instructions[0xEB] = shortJump
	instructions[0xFF] = codeFF
}

func shortJump(emu *Emulator) {
	diff := emu.getSignCode8(1)
	emu.eip += uint32(diff + 2)
}

func nearJump(emu *Emulator) {
	diff := emu.getSignCode32(1)
	emu.eip += uint32(diff + 5)
}

func movR32Imm32(emu *Emulator) {
	reg := emu.getCode8(0) - 0xB8
	val := emu.getCode32(1)
	emu.registers[reg] = val
	emu.eip += 5
}

func movRm32Imm32(emu *Emulator) {
	emu.eip++
	modrm := ParseModRM(emu)
	val := emu.getCode32(0)
	emu.eip += 4
	emu.setRm32(modrm, val)
}

func movRm32R32(emu *Emulator) {
	emu.eip++
	modrm := ParseModRM(emu)
	r32 := emu.getRegister32(modrm.regIndex)
	emu.setRm32(modrm, r32)
}

func movR32Rm32(emu *Emulator) {
	emu.eip++
	modrm := ParseModRM(emu)
	rm32 := emu.getRm32(modrm)
	emu.setRegister32(modrm.regIndex, rm32)
}

func addRm32R32(emu *Emulator) {
	emu.eip++
	modrm := ParseModRM(emu)
	rm32 := emu.getRm32(modrm)
	r32 := emu.getRegister32(modrm.regIndex)
	emu.setRm32(modrm, rm32+r32)
}

func addRm32Imm8(emu *Emulator, modrm *ModRM) {
	rm32 := emu.getRm32(modrm)
	imm8 := emu.getSignCode8(0)
	emu.eip++
	emu.setRm32(modrm, rm32+uint32(imm8))
}

func code83(emu *Emulator) {
	emu.eip++
	modrm := ParseModRM(emu)

	switch modrm.opcode {
	case 0:
		addRm32Imm8(emu, modrm)
	case 5:
		subRm32Imm8(emu, modrm)
	case 7:
		cmpRm32Imm8(emu, modrm)
	default:
		log.Printf("not implemented: 83 /%d\n", modrm.opcode)
		os.Exit(1)
	}
}

func subRm32Imm8(emu *Emulator, modrm *ModRM) {
	rm32 := emu.getRm32(modrm)
	imm8 := uint32(emu.getSignCode8(0))
	emu.eip++
	emu.setRm32(modrm, rm32-imm8)
	emu.updateEflagsSub(rm32, imm8)
}

func codeFF(emu *Emulator) {
	emu.eip++
	modrm := ParseModRM(emu)

	switch modrm.opcode {
	case 0:
		incRm32(emu, modrm)
	default:
		log.Printf("not implemented: FF /%d\n", modrm.opcode)
		os.Exit(1)
	}
}

func incRm32(emu *Emulator, modrm *ModRM) {
	val := emu.getRm32(modrm)
	emu.setRm32(modrm, val+1)
}

func pushR32(emu *Emulator) {
	reg := emu.getCode8(0) - 0x50
	emu.push32(emu.getRegister32(reg))
	emu.eip++
}

func pushImm32(emu *Emulator) {
	val := emu.getCode32(1)
	emu.push32(val)
	emu.eip += 5
}

func pushImm8(emu *Emulator) {
	val := emu.getCode8(1)
	emu.push32(uint32(val))
	emu.eip += 2

}

func popR32(emu *Emulator) {
	reg := emu.getCode8(0) - 0x58
	emu.setRegister32(reg, emu.pop32())
	emu.eip++
}

func callRel32(emu *Emulator) {
	diff := emu.getSignCode32(1)
	emu.push32(emu.eip + 5)
	emu.eip += uint32(diff + 5)
}

func ret(emu *Emulator) {
	emu.eip = emu.pop32()
}

func leave(emu *Emulator) {
	ebp := emu.getRegister32(uint8(EBP))
	emu.setRegister32(uint8(ESP), ebp)
	emu.setRegister32(uint8(EBP), emu.pop32())
	emu.eip++
}

func cmpR32Rm32(emu *Emulator) {
	emu.eip++
	modrm := ParseModRM(emu)
	r32 := emu.getRegister32(modrm.regIndex)
	rm32 := emu.getRm32(modrm)
	emu.updateEflagsSub(r32, rm32)
}

func cmpRm32Imm8(emu *Emulator, modrm *ModRM) {
	rm32 := emu.getRm32(modrm)
	imm8 := uint32(emu.getCode8(0))
	emu.eip++
	emu.updateEflagsSub(rm32, imm8)
}

func jc(emu *Emulator) {
	var diff uint32
	if emu.isCarry() {
		diff = uint32(emu.getSignCode8(1))
	} else {
		diff = 0
	}
	emu.eip += diff + 2
}

func jz(emu *Emulator) {
	var diff uint32
	if emu.isZero() {
		diff = uint32(emu.getSignCode8(1))
	} else {
		diff = 0
	}
	emu.eip += diff + 2
}

func js(emu *Emulator) {
	var diff uint32
	if emu.isSign() {
		diff = uint32(emu.getSignCode8(1))
	} else {
		diff = 0
	}
	emu.eip += diff + 2
}

func jo(emu *Emulator) {
	var diff uint32
	if emu.isOverflow() {
		diff = uint32(emu.getSignCode8(1))
	} else {
		diff = 0
	}
	emu.eip += diff + 2
}

func jl(emu *Emulator) {
	var diff uint32
	if emu.isSign() != emu.isOverflow() {
		diff = uint32(emu.getSignCode8(1))
	} else {
		diff = 0
	}
	emu.eip += diff + 2
}

func jle(emu *Emulator) {
	var diff uint32
	if emu.isZero() || (emu.isSign() != emu.isOverflow()) {
		diff = uint32(emu.getSignCode8(1))
	} else {
		diff = 0
	}
	emu.eip += diff + 2
}

func jnc(emu *Emulator) {
	var diff uint32
	if emu.isCarry() {
		diff = 0
	} else {
		diff = uint32(emu.getSignCode8(1))
	}
	emu.eip += diff + 2
}

func jnz(emu *Emulator) {
	var diff uint32
	if emu.isZero() {
		diff = 0
	} else {
		diff = uint32(emu.getSignCode8(1))
	}
	emu.eip += diff + 2
}

func jns(emu *Emulator) {
	var diff uint32
	if emu.isSign() {
		diff = 0
	} else {
		diff = uint32(emu.getSignCode8(1))
	}
	emu.eip += diff + 2
}

func jno(emu *Emulator) {
	var diff uint32
	if emu.isOverflow() {
		diff = 0
	} else {
		diff = uint32(emu.getSignCode8(1))
	}
	emu.eip += diff + 2
}
