package main

import (
	"fmt"
	"time"
)

func main() {
	sendingInterval := time.Second
	receptionInterval := time.Second * 2
	intChan := make(chan int, 0)
	go func() {
		var ts0, ts1 int64
		for i := 1; i <= 5; i++ {
			intChan <- i
			ts1 = time.Now().Unix()
			if ts0 == 0 {
				fmt.Println("Sent:", i)
			} else {
				fmt.Printf("Sent: %d [interval: %d s]\n", i, ts1-ts0)
			}
			ts0 = time.Now().Unix()
			time.Sleep(sendingInterval)
		}
		close(intChan)
	}()
	var ts0, ts1 int64
Loop:
	for {
		select {
		case v, ok := <-intChan:
			if !ok {
				break Loop
			}
			ts1 = time.Now().Unix()
			if ts0 == 0 {
				fmt.Println("Received:", v)
			} else {
				fmt.Printf("Received: %d [interval: %d s]\n", v, ts1-ts0)
			}
		}
		ts0 = time.Now().Unix()
		time.Sleep(receptionInterval)
	}
	fmt.Println("End.")
}

/*
Sent: 1
Received: 1
Sent: 2 [interval: 2 s]
Received: 2 [interval: 2 s]
Received: 3 [interval: 2 s]
Sent: 3 [interval: 2 s]
Received: 4 [interval: 2 s]
Sent: 4 [interval: 2 s]
Received: 5 [interval: 2 s]
Sent: 5 [interval: 3 s]
End.
*/

/*
intChan 改为容量为5，出现如下结果, 即便通道有缓冲时，向intChan发送值，另一边会立刻收到
Sent: 1
Received: 1
Sent: 2 [interval: 1 s]
Received: 2 [interval: 2 s]
Sent: 3 [interval: 1 s]
Sent: 4 [interval: 1 s]
Received: 3 [interval: 2 s]
Sent: 5 [interval: 1 s]
Received: 4 [interval: 2 s]
Received: 5 [interval: 2 s]
End.
*/
