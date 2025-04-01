package mnemonic

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// 硬化索引的基值（2^31）
const HardenedOffset = 0x80000000

// 以太坊 BIP-44 派生路径："m/44'/60'/0'/0/index"
var ethPath = []uint32{44 + HardenedOffset, 60 + HardenedOffset, 0 + HardenedOffset, 0}

type Mnemonic struct {
	words     string
	masterKey *bip32.Key
}

func NewMnemonic(words string) (*Mnemonic, error) {
	// 生成种子
	seed := bip39.NewSeed(words, "")
	// 生成主密钥
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("创建主密钥失败: %v", err)
	}
	return &Mnemonic{words: words, masterKey: masterKey}, nil
}

func (m *Mnemonic) GetAddressAndPrivateKeyByIndex(index int) (string, string, error) {
	address, privateKey, err := deriveEthereumAddress(m.masterKey, index)
	if err != nil {
		return "", "", err
	}
	return address, privateKey, nil
}

func (m *Mnemonic) GetAddressByIndex(index int) (string, error) {
	address, _, err := deriveEthereumAddress(m.masterKey, index)
	if err != nil {
		return "", err
	}
	return address, nil
}

func (m *Mnemonic) GetPrivateKeyByIndex(index int) (string, error) {
	_, privateKey, err := deriveEthereumAddress(m.masterKey, index)
	if err != nil {
		return "", err
	}
	return privateKey, nil
}

// 通过索引派生以太坊地址
func deriveEthereumAddress(masterKey *bip32.Key, index int) (string, string, error) {
	// 遍历 BIP-44 以太坊路径
	key := masterKey
	for _, idx := range ethPath {
		var err error
		key, err = key.NewChildKey(idx)
		if err != nil {
			return "", "", fmt.Errorf("派生路径失败: %v", err)
		}
	}

	// 派生最终的子密钥（索引位置）
	childKey, err := key.NewChildKey(uint32(index))
	if err != nil {
		return "", "", fmt.Errorf("派生地址索引失败: %v", err)
	}

	// 转换为 ECDSA 私钥
	privateKey, err := crypto.ToECDSA(childKey.Key)
	if err != nil {
		return "", "", fmt.Errorf("转换私钥失败: %v", err)
	}

	// 计算以太坊地址
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// 返回以太坊地址和私钥
	return address.Hex(), hexutil.Encode(crypto.FromECDSA(privateKey)), nil
}
