package main

import (
	"hash/crc32"
)

type Packet struct {
	Type    byte   `json:type`
	Content []byte `json:content`
}

func agreementData(transBytes []byte) []byte {
	length := len(transBytes) + 8
	result := make([]byte, length)
	result[0] = 0xFF
	result[1] = 0xFF
	result[2] = byte(uint16(len(transBytes)) >> 8)
	result[3] = byte(uint16(len(transBytes)) & 0xFF)
	copy(result[4:], transBytes)
	sendCrc := crc32.ChecksumIEEE(transBytes)
	result[length-4] = byte(sendCrc >> 24)
	result[length-3] = byte(sendCrc >> 16 & 0xFF)
	result[length-2] = 0xFF
	result[length-1] = 0xFE
	return result
}
