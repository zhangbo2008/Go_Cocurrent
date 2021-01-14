package main  //这个例子来进行客户端ctx , 设置超时判断. 超时就立即发送给客户.

import (
"context"
"fmt"
"io/ioutil"
"net/http"
"sync"
"time"
)

// 客户端

type respData struct {
	resp *http.Response
	err  error
}

func doCall(ctx context.Context) {
	transport := http.Transport{
		// 请求频繁可定义全局的client对象并启用长链接
		// 请求不频繁使用短链接
		DisableKeepAlives: true, 	}  //建立一个长时间的链接.
	client := http.Client{
		Transport: &transport,          //建立请求
	}

	respChan := make(chan *respData, 1)  //返回的东西放入这个cahnnel
	req, err := http.NewRequest("GET", "http://127.0.0.1:8000/", nil)
	if err != nil {
		fmt.Printf("new requestg failed, err:%v\n", err)
		return
	}
	req = req.WithContext(ctx) // 使用带超时的ctx创建一个新的client request , req进行包装一层ctx
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait()   // 因为client只发送一次,所以需要用WaitGroup来卡主主进程.
	go func() {
		resp, err := client.Do(req)
		fmt.Printf("client.do resp:%v, err:%v\n", resp, err)
		rd := &respData{
			resp: resp,
			err:  err,
		}
		respChan<- rd //返回结果进入channel
		wg.Done()
	}()

	select {
	case<-ctx.Done():  // ctx.done表示超时了.
		//transport.CancelRequest(req)
		fmt.Println("call api timeout")
	case result :=<-respChan:
		fmt.Println("call server api success")
		if result.err != nil {
			fmt.Printf("call server api failed, err:%v\n", result.err)
			return
		}
		defer result.resp.Body.Close()
		data, _ := ioutil.ReadAll(result.resp.Body)
		fmt.Printf("resp:%v\n", string(data))
	}
}

func main() {
	// 定义一个100毫秒的超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel() // 调用cancel释放子goroutine资源, 运行完之后,调用cancel来关闭资源.
	doCall(ctx)
}
