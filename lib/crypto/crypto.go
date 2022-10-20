package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
)

type Decoder interface {
	Decode(key []byte) Encoder
	ToBytes() []byte
	SetBytes(data []byte) Decoder
}

type Encoder interface {
	Encode(key []byte) Decoder
	ToString() string
	SetString(data string) Encoder
}

func Sha1Str(text string) string {
	t := sha1.New()
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
