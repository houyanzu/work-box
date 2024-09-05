package loop

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Deprecated: Use Loop2 instead.
func Loop(t int, fs ...func(wg *sync.WaitGroup)) {
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	fmt.Println("程序已启动...")
	wg := &sync.WaitGroup{}

	go func() {
		for {
			fmt.Println("start")
			for _, f := range fs {
				wg.Add(1)
				f(wg)
			}
			fmt.Println("ok")
			time.Sleep(time.Duration(t) * time.Second)
		}
	}()

	<-c
	fmt.Println("正在退出，请稍后...")
	wg.Wait()
	fmt.Println("退出成功！")
	os.Exit(0)
}

func Loop2(t int, fs ...func()) {
	c := make(chan os.Signal, 1) // 加入缓冲区，防止信号丢失或阻塞
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	fmt.Println("程序已启动...")
	wg := &sync.WaitGroup{}

	// 控制协程退出的标志
	exit := make(chan struct{})

	go func() {
		for {
			select {
			case <-exit:
				return // 收到退出信号，退出循环
			default:
				fmt.Println("start")
				for _, f := range fs {
					wg.Add(1)
					func(f func()) {
						defer wg.Done()
						f()
					}(f)
				}
				fmt.Println("ok")
				time.Sleep(time.Duration(t) * time.Second)
			}
		}
	}()

	<-c // 等待系统信号
	fmt.Println("正在退出，请稍后...")

	// 通知 goroutine 退出循环
	close(exit)

	wg.Wait() // 等待所有 goroutine 结束
	fmt.Println("退出成功！")
	signal.Stop(c) // 停止信号监听
	os.Exit(0)
}
