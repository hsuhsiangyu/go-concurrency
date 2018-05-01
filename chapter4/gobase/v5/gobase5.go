package main

import (
	"fmt"
	"time"
)

func main() {
	names := []string{"Eric", "Harry", "Robert", "Jim", "Mark"}
	for _, name := range names {
		go func(who string) { // name 值被复制并在go函数中由参数who指代
			fmt.Printf("Hello, %s!\n", who)
		}(name) // go函数也可以有参数声明，解决了v4的问题
	}
	time.Sleep(time.Millisecond)
}
