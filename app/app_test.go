package app

import (
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"reflect"
	"testing"
)

func TestTron(t *testing.T) {
	addr := address.HexToAddress("0x41b7E6543D10f192dBD83A285FFa074Dbf37a541E3")
	fmt.Println(addr.String())
}

func TestReflect(t *testing.T) {
	var userID interface{}
	userID = 1
	tt := reflect.TypeOf(userID)
	fmt.Println(tt)
}
