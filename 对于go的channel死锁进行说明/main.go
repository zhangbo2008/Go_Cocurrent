package main

import (
	"fmt"
	"time"
)








func testDeadLock(c chan int){
	go func()   {
		for {
			fmt.Println(<-c)
		}
	}()
}






func test(c chan int){
	c<-'A'
	c<-'A'
	c<-'A'
	c<-'A'
	c<-'A'
}


func main(){
/*
解释一下为什么一定要写go,
因为一个channel他如果没有缓冲那么他里面写入了,他就会卡死,然后直到别人读取他才会继续走.
如果go test(c) 这行不写go那么运行完test就直接卡死了.main这个线程就卡死.导致整个程序卡死.

如果写了go那么go后面的代码就会起一个线程来跑,而不再main里面跑.所以main还是可以继续后面的代码运行.
同理 下面的testDeadLock 函数也需要写go.


*/
	c :=make(chan int )  //跟channel有关的一定要放到读写里面 go,不然他就会卡死
	go test(c)

	go testDeadLock(c)
	time.Sleep(2000) //保证子进程都运行结束才关闭main







	 //var a=	NewDispatcher(2);
	 //a.Run()




////启动时候往jobqueue里面扔东西就行了.

	//
	//for i := 0; i < 10; i++ {
	//	JobQueue<-Job{}
	//}
	//
	//



	//fmt.Println(reflect.TypeOf(a))








}