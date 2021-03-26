package main

import (
	"fmt"
)

var (
	biosToTerminal [8]int = [8]int{30, 34, 32, 36, 31, 35, 33, 37}
)

func putString(s string) {
	for _, v := range s {
		ioOut8(0x03f8, uint8(v))
	}
}

func biosVideoTeletype(emu *Emulator) {
	col := emu.getRegister8(BL) & 0x0f
	ch := emu.getRegister8(AL)
	terminalColor := biosToTerminal[col&0x07]
	bright := (col & 0x08) >> 3
	s := fmt.Sprintf("\x1b[%d;%dm%c\x1b[0m", bright, terminalColor, ch)
	putString(s)
}

func biosVideo(emu *Emulator) {
	fn := emu.getRegister8(AH)
	switch fn {
	case 0x0e:
		biosVideoTeletype(emu)
	default:
		fmt.Printf("not implemented BIOS video function: 0x%02x\n", fn)
	}
}
