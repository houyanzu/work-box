package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

type Decoder interface {
	Decode(key []byte) string
	ToBytes() []byte
	SetBytes(data []byte)
}

type Encoder interface {
	Encode(key []byte) []byte
	ToString() string
	SetString(data string)
}

func Sha1Str(text string) string {
	t := sha1.New()
	_, _ = io.WriteString(t, text)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func Sha256Str(text string) string {
	t := sha256.New()
	_, _ = io.WriteString(t, text)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func Md5Str(text string) string {
	t := md5.New()
	_, _ = io.WriteString(t, text)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func Base64StrEncode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

func Base64StrDecode(text string) string {
	res, _ := base64.StdEncoding.DecodeString(text)
	return string(res)
}

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
