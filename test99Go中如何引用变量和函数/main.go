//包名
package main
//导入test包
import(
	"awesomeProject8/test99Go中如何引用变量和函数/test"

)
func main() {
	// 调用 test 内的 Println 方法
	var b=test.A // go里面引入其他文件变量的时候需要大写才行.
	print(b)

}