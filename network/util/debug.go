package util

import (
	"bytes"
	"fmt"
	"net"
)

func ApplicationHexDump(data []byte) {
	fmt.Printf(" L7 (Payload)  : %d bytes\n", len(data))
	fmt.Println("-----------------------------------------------------------------")
	const bytesPerLine = 16
	var hexBuf, asciiBuf bytes.Buffer

	for i := 0; i < len(data); i += bytesPerLine {
		fmt.Printf("  %08x: ", i)

		hexBuf.Reset()
		asciiBuf.Reset()

		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}

		line := data[i:end]

		for j := 0; j < len(line); j++ {
			hexBuf.WriteString(fmt.Sprintf("%02x ", line[j]))
			if j == 7 {
				hexBuf.WriteString(" ")
			}
		}

		for _, b := range line {
			if b >= 32 && b <= 126 {
				asciiBuf.WriteByte(b)
			} else {
				asciiBuf.WriteByte('.')
			}
		}

		hexStr := hexBuf.String()
		padding := (bytesPerLine * 3) + (bytesPerLine / 8)
		fmt.Printf("%-*s |%s|\n", padding, hexStr, asciiBuf.String())
	}
	fmt.Println("-----------------------------------------------------------------")
}

func IsLocalIP(ipStr string) bool {
	if ipStr == "127.0.0.1" || ipStr == "::1" || ipStr == "localhost" {
		return true
	}

	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, addr := range interfaces {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip != nil && ip.String() == ipStr {
			return true
		}
	}
	return false
}
