package main

import (
	"time"
)

// go 并发https://blog.csdn.net/lein_wang/article/details/82746339
//http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/
/*

所有的worker共享job queue











*/








// 先定义所有的类型!!!!!!!!!!!!!
type Worker struct {
	WorkerPool  chan chan Job
	JobChannel JobChan  //工人维护一个job队列.
	quit       chan bool
}





type Job struct {
	a int
}


func (w *Job) Do() {
	println("服务了一个",w.a)

}

var abc= Job{}

// A buffered channel that we can send work requests on.


// define job channel
type JobChan chan Job

// define worker channer
type WorkerChan chan JobChan   //所有的worker组成worker 队列.


var (
	//JobQueue          chan Job
	JobQueue=make(chan  Job, 200)         //为什么用上面的不好使???????????? 无缓冲的不好使why?
)






//再定义所有的方法.
func (w *Worker) Start() {
	go func() {
		//print(333333333333333333)
		for {
			// regist current job channel to worker pool  // debug技巧: 看channel里面的qcount 里面的数量就是channel里面现在存在的东西数量.
			w.WorkerPool <- w.JobChannel     //抽象含义表示这个channel已经用完了,这个worker又可以回到pool中了.
//print(44444444444444)
			select {

			case job := <-w.JobChannel:
				//print(5555555555)
				//print("worker  job := <-w.JobChannel")
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


func NewDispatcher(maxWorkers int) *Dispatcher {  //go没法传默认参数,曹乐!
	pool := make(chan chan Job, maxWorkers)
	//println("简历时候的workerpool",&pool)
	return &Dispatcher{WorkerPool: pool}//这里面简历的workerpool表示大家共用的了.
}


func NewWorker(workerPool chan chan Job) Worker {
	//println("看看Workerpool的地址2",&workerPool,)
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}
func (d *Dispatcher) Run() {
	//println("看看Workerpool的地址1",&d.WorkerPool,)
	for i := 0; i < 2; i++ {
		//println("简历newworker时候的workerpool",&d.WorkerPool)
		worker := NewWorker(d.WorkerPool) //这里面保证了所有的NewWorker共享了d.WorkerPool这个全局变量.


		//println("建荔湾worker之后的workerpool",&worker.WorkerPool)
		worker.Start()
	}
	go d.dispatch()  //启动派遣函数.只要设计了<-运算就要进行go修饰.
	//for {
	//	 job := <-JobQueue
	//		go func(job Job) {
	//			print("111111111121212")
	//			jobChan := <-WorkerPool
	//			jobChan <- job
	//		}(job)
	//	// stop dispatcher
	//
	//	}
}

func (d *Dispatcher) dispatch() {
	//print(99999999)
	//print(&JobQueue,34234234)
	for {
		select {
		case job := <-JobQueue:// 这步没跑
			//print("---------")
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				//print(77777777)  // 下面这一行没走!!!!!!!!!!!!!
				jobChannel := <-d.WorkerPool  //抽象含义表示获取一个worker   jobChannel表示一个worker
				//print(23123)
				// dispatch the job to the worker job channel
				//print(6666666666)
				jobChannel <- job   //然后把job 给worker --jobChannel里面有了就会激活等待的worker start代码里面的进程
			}(job)
		}
	}
}








func main(){







	  a:=	NewDispatcher(2);
	  go a.Run()           //先启动消费者




////启动时候往jobqueue里面扔东西就行了.

	//  需要把channel操作扔到go里面封装来避免死锁.
	go func() 	{
		for i := 0; i < 10; i++ {
			//print("insert")
			//print(&JobQueue,34234234)
			//a:=Job{343434}
			//JobQueue <- Job{} //这行没启动????????????????????????为什么阻塞了!!!!!!!!!!
			work := Job{i}
//print("^^^^^^^^^^^^^^^^^^^^^^6")
			// Push the work onto the queue.
			JobQueue <- work
			//print(101010)
		}
	}()
	  print("main-----------------")




	//fmt.Println(reflect.TypeOf(a))

// 最后循环打印,来保证整个程序的运行,防止主进程直接退出导致子进程没跑完.

		time.Sleep(200000 * time.Second)






}