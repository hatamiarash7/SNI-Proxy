package main

import "fmt"

var TLSHeaderLength = 5

func getHost(data []byte) (string, error) {
	if len(data) == 0 || data[0] != 0x16 {
		return "", fmt.Errorf("Doesn't look like a TLS Client Hello")
	}

	extensions, err := getExtBlock(data)

	if err != nil {
		return "", err
	}

	sn, err := getSNBlock(extensions)

	if err != nil {
		return "", err
	}

	sni, err := getSNIBlock(sn)

	if err != nil {
		return "", err
	}

	return string(sni), nil
}

func dataLength(data []byte, index int) int {
	b1 := int(data[index])
	b2 := int(data[index+1])

	return (b1 << 8) + b2
}

func getSNIBlock(data []byte) ([]byte, error) {
	index := 0

	for {
		if index >= len(data) {
			break
		}

		length := dataLength(data, index)
		endIndex := index + 2 + length

		if data[index+2] == 0x00 {
			sni := data[index+3:]
			sniLength := dataLength(sni, 0)

			return sni[2 : sniLength+2], nil
		}

		index = endIndex
	}

	return []byte{}, fmt.Errorf(
		"There is not any SNI in SN block",
	)
}

func getSNBlock(data []byte) ([]byte, error) {
	index := 0

	if len(data) < 2 {
		return []byte{}, fmt.Errorf("Not enough bytes for SN block")
	}

	extensionLength := dataLength(data, index)

	if extensionLength+2 > len(data) {
		return []byte{}, fmt.Errorf("Extension looks not good")
	}

	data = data[2 : extensionLength+2]

	for {
		if index+4 >= len(data) {
			break
		}

		length := dataLength(data, index+2)
		endIndex := index + 4 + length

		if data[index] == 0x00 && data[index+1] == 0x00 {
			return data[index+4 : endIndex], nil
		}

		index = endIndex
	}

	return []byte{}, fmt.Errorf(
		"There is not any SNI in Extension block",
	)
}

func getExtBlock(data []byte) ([]byte, error) {
	var index = TLSHeaderLength + 38

	if len(data) <= index+1 {
		return []byte{}, fmt.Errorf("Not enough bytes for Client Hello")
	}

	if newIndex := index + 1 + int(data[index]); (newIndex + 2) < len(data) {
		index = newIndex
	} else {
		return []byte{}, fmt.Errorf("Not enough bytes for SessionID")
	}

	if newIndex := (index + 2 + dataLength(data, index)); (newIndex + 1) < len(data) {
		index = newIndex
	} else {
		return []byte{}, fmt.Errorf("Not enough bytes for Cipher List")
	}

	if newIndex := index + 1 + int(data[index]); newIndex < len(data) {
		index = newIndex
	} else {
		return []byte{}, fmt.Errorf("Not enough bytes for compression length")
	}

	if len(data[index:]) == 0 {
		return nil, fmt.Errorf("No extension")
	}

	return data[index:], nil
}