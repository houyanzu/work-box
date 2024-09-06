package boxlog

import (
	"log"
	"testing"
)

func TestLog(t *testing.T) {
	err := Init("testlog", true)
	if err != nil {
		log.Fatal(err)
	}
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")
	Debug("debug,%s", "test")
	Info("info, %d", 123)
	Warn("warn")
	Error("error")
}
