package main

type ModRM struct {
	mod      uint8
	opcode   uint8
	regIndex uint8
	rm       uint8
	sib      uint8
	disp8    int8
	disp32   uint32
}

func ParseModRM(emu *Emulator) *ModRM {
	modrm := new(ModRM)

	code := emu.getCode8(0)
	modrm.mod = code >> 6
	modrm.opcode, modrm.regIndex = (code&0x38)>>3, (code&0x38)>>3
	modrm.rm = code & 0x07

	emu.eip++

	if modrm.mod != 3 && modrm.rm == 4 {
		modrm.sib = emu.getCode8(0)
		emu.eip++
	}

	if (modrm.mod == 0 && modrm.rm == 5) || modrm.mod == 2 {
		modrm.disp8, modrm.disp32 = int8(emu.getCode32(0)), emu.getCode32(0)
		emu.eip += 4
	} else if modrm.mod == 1 {
		modrm.disp8, modrm.disp32 = emu.getSignCode8(0), uint32(emu.getCode8(0))
		emu.eip++
	}

	return modrm
}
