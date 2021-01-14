package main
// 这个方法是用channel来设置信号.来停止并发.
import (
	"fmt"
	"sync"

	"time"
)

var wg sync.WaitGroup //信号量保证主进程不会停止.

// 管道方式存在的问题：
// 1. 使用全局变量在跨包调用时不容易实现规范和统一，需要维护一个共用的channel

func worker(exitChan chan struct{}) {

 LOOP:
	for {
		fmt.Println("worker")
		time.Sleep(time.Second)
		select {
		case <-exitChan: // 等待接收上级通知
			break LOOP  //使用这个跳出外层循环.

		default:
		}
	}
	wg.Done()
}

func main() {
	var exitChan = make(chan struct{})
	wg.Add(1)
	go worker(exitChan)
	time.Sleep(time.Second * 3) // sleep3秒以免程序过快退出
	exitChan <- struct{}{}      // 给子goroutine发送退出信号
	//close(exitChan)
	wg.Wait()
	fmt.Println("over")
}