package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(2 * time.Second)
	fmt.Printf("Present time: %v.\n", time.Now())
	expirationTime := <-timer.C  // 在time.Timer类型中，对外通知定时器到期的途径就是通道，由字段C表示，一直阻塞，直到定时器到期
	fmt.Printf("Expiration time: %v.\n", expirationTime)
	fmt.Printf("Stop timer: %v.\n", timer.Stop())
}
