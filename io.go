package main

import "fmt"

func ioIn8(address uint16) uint8 {
	switch address {
	case 0x03f8:
		var s string
		fmt.Scanln(&s)
		return s[0]
	default:
		return 0
	}
}

func ioOut8(address uint16, val uint8) {
	switch address {
	case 0x03f8:
		fmt.Printf("%c", val)
	}
}
