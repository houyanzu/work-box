package loop

import (
	"fmt"
	"testing"
)

func TestLoop2(t *testing.T) {
	Loop2(1, func() {
		fmt.Println(123)
	})
}
