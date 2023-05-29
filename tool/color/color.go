package color

import (
	"fmt"
)

var (
	//kernel32    *syscall.LazyDLL  = syscall.NewLazyDLL(`kernel32.dll`)
	//proc        *syscall.LazyProc = kernel32.NewProc(`SetConsoleTextAttribute`)
	//CloseHandle *syscall.LazyProc = kernel32.NewProc(`CloseHandle`)

	// 给字体颜色对象赋值
	FontColor Color = Color{30, 31, 32, 32, 33, 34, 35, 36, 37}
)

type Color struct {
	Black       int //
	Red         int //
	Green       int //
	Light_green int //
	Yellow      int //
	Blue        int //
	Purple      int //
	Cyan        int //
	Light_gray  int //
}

// 输出有颜色的字体
func ColorPrint(s string, i int) {
	fmt.Printf(" %c[%d;40;%dm%s%s%c[0m\n ", 0x1B, 0, i, "", s, 0x1B)
}
