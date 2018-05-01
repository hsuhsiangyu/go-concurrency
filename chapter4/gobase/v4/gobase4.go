package main

import (
	"fmt"
	"time"
)

func main() {
	names := []string{"Eric", "Harry", "Robert", "Jim", "Mark"}
	for _, name := range names {  // 并发执行的5个go 函数，它们都是在for语句执行完毕之后才执行的，而name在这时指代的值已经是Mark了
		go func() {
			fmt.Printf("Hello, %s!\n", name)
		}()
	}  // 打印出5行 Hello, Mark!
	time.Sleep(time.Millisecond)
}
