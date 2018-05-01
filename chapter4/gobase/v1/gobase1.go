package main

func main() {
	go println("Go! Goroutine!")  //不会出现Go! Goroutine, main函数结束，该G还没来得及执行
}
