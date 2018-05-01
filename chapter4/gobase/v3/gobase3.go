package main

import (
	"fmt"
	"time"
)

func main() {
	name := "Eric"
	go func() {
		fmt.Printf("Hello, %s!\n", name)
	}()
	//name = "Harry"
	//time.Sleep(time.Millisecond)  // 这句放在后面，一般会打印出 Hello, Harry!
	time.Sleep(time.Millisecond)  // 这句放在前面，一般会打印出 Hello, Eric
     name = "Harry"
}
