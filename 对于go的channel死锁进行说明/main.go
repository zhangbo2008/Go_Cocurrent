package main

import (
	"fmt"
	"time"
)

// go 并发https://blog.csdn.net/lein_wang/article/details/82746339
//http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/
/*

所有的worker共享job queue











*/








// 先定义所有的类型!!!!!!!!!!!!!
type Worker struct {
	JobChannel JobChan  //工人维护一个job队列.
	quit       chan bool
}





type Job struct {
	a int
}


func (w *Job) Do() {
	print(1111111111)

}

var abc= Job{}

// A buffered channel that we can send work requests on.


// define job channel
type JobChan chan Job

// define worker channer
type WorkerChan chan JobChan   //所有的worker组成worker 队列.


var (
	JobQueue          JobChan
	WorkerPool        WorkerChan
)






//再定义所有的方法.
func (w *Worker) Start() {
	go func() {
		for {
			// regist current job channel to worker pool
			WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				job.Do()

			// recieve quit event, stop worker
			case <-w.quit:
				return
			}
		}
	}()
}


//Dispatcher 是主控制器.  workerPool 是全部的核心.
type Dispatcher struct {
	WorkerPool chan chan Job
}


func NewDispatcher(maxWorkers int) Dispatcher {  //go没法传默认参数,曹乐!
	pool := make(chan chan Job, maxWorkers)
	return Dispatcher{WorkerPool: pool}
}


func NewWorker(workerPool chan chan Job) Worker {
	return Worker{

		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}
func (d *Dispatcher) Run() {
	for i := 0; i < 10; i++ {
		worker := NewWorker(d.WorkerPool)

		worker.Start()
	}

	for {
		select {
		case job := <-JobQueue:
			go func(job Job) {
				jobChan := <-WorkerPool
				jobChan <- job
			}(job)
		// stop dispatcher

		}
	}
}







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
	c :=make(chan int)  //跟channel有关的一定要放到读写里面,不然他就会卡死
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