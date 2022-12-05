package app

import (
	"fmt"
	"github.com/houyanzu/work-box/lib/httptool"
	"testing"
)

func TestTron(t *testing.T) {
	resp, code, err := httptool.Get("https://api.shasta.trongrid.io/wallet/getnowblock")
	if err != nil {
		panic(err)
	}
	fmt.Println(code, string(resp))
}
