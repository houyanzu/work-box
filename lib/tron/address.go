package tron

import (
	"errors"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/houyanzu/work-box/tool/eth"
)

func HexToTronAddress(hex string) (string, error) {
	if !eth.IsAddress(hex) {
		return "", errors.New("wrong hex")
	}

	hex = "0x41" + hex[2:]
	a := address.HexToAddress(hex)
	return a.String(), nil
}

func TronAddressToHex(tronAddress string) (string, error) {
	a, err := address.Base58ToAddress(tronAddress)
	if err != nil {
		return "", err
	}
	return "0x" + a.Hex()[4:], nil
}
