package main

import (
	"bytes"
	"fmt"
	"moul.io/http2curl"
	"net/http"
)
//var a =3;
func main(){
//print(a)

	data := bytes.NewBufferString(`{"hello":"world","answer":42}`)
	req, _ := http.NewRequest("PUT", "http://www.example.com/abc/def.ghi?jlk=mno&pqr=stu", data)
	req.Header.Set("Content-Type", "application/json")
//设置好requeset, 然后进行函数转化即可.
	command, _ := http2curl.GetCurlCommand(req)
	fmt.Println(command)
}
