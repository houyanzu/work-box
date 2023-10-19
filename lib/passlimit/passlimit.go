package passlimit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxAttempts = 3
const lockoutDuration = 5 * time.Minute
const stateFileName = "pass_state.enc"

var mu sync.Mutex

// 加密和解密所需的密钥
var encryptionKey = []byte("super-houjia-key")

// 加密状态信息
func encryptState(attempts int, lockedUntil time.Time) (string, error) {
	plaintext := fmt.Sprintf("%d:%s", attempts, lockedUntil.Format(time.RFC3339))
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return hex.EncodeToString(ciphertext), nil
}

// 解密状态信息
func decryptState(ciphertextHex string) (int, time.Time, error) {
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return 0, time.Time{}, err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return 0, time.Time{}, err
	}

	if len(ciphertext) < aes.BlockSize {
		return 0, time.Time{}, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	parts := string(ciphertext)
	split := func(c rune) bool { return c == ':' }
	tokens := strings.FieldsFunc(parts, split)

	if len(tokens) != 2 {
		return 0, time.Time{}, errors.New("invalid ciphertext format")
	}

	attempts, err := strconv.Atoi(tokens[0])
	if err != nil {
		return 0, time.Time{}, err
	}

	lockedUntil, err := time.Parse(time.RFC3339, tokens[1])
	if err != nil {
		return 0, time.Time{}, err
	}

	return attempts, lockedUntil, nil
}

// 保存状态信息到文件
func saveState(attempts int, lockedUntil time.Time) error {
	ciphertext, err := encryptState(attempts, lockedUntil)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	file, err := os.Create(stateFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(ciphertext)
	return err
}

// 从文件加载状态信息
func loadState() (int, time.Time, error) {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(stateFileName)
	if err != nil {
		return 0, time.Time{}, err
	}
	defer file.Close()

	var ciphertextHex string
	_, err = fmt.Fscanf(file, "%s", &ciphertextHex)
	if err != nil {
		return 0, time.Time{}, err
	}

	attempts, lockedUntil, err := decryptState(ciphertextHex)
	return attempts, lockedUntil, err
}

// 尝试登录并更新状态
func TryPass(isPass bool) (bool, error) {
	attempts, lockedUntil, err := loadState()
	if err != nil {
		return false, err
	}

	if lockedUntil.After(time.Now()) {
		return false, nil
	}

	// 在这里执行密码验证逻辑，如果验证成功返回 true，否则返回 false
	if isPass {
		// 登录成功，重置尝试次数
		attempts = 0
		lockedUntil = time.Time{}
	} else {
		// 登录失败，增加尝试次数
		attempts++
		if attempts >= maxAttempts {
			lockedUntil = time.Now().Add(lockoutDuration)
		}
	}

	// 保存状态信息
	err = saveState(attempts, lockedUntil)
	if err != nil {
		return false, err
	}

	return attempts == 0, nil
}
