package main

import (
	"fmt"
	"time"
)

var strChan = make(chan string, 3)

func main() {
	syncChan1 := make(chan struct{}, 1)  // 用于传递信号的通道都以struc{}作为元素类型
	syncChan2 := make(chan struct{}, 2)
	go func() { // 用于演示接收操作。
		<-syncChan1
		fmt.Println("Received a sync signal and wait a second... [receiver]")
		time.Sleep(time.Second)
		for {
			if elem, ok := <-strChan; ok {
				fmt.Println("Received:", elem, "[receiver]")
			} else {
				break
			}
		}
		fmt.Println("Stopped. [receiver]")
		syncChan2 <- struct{}{}
	}()
	go func() { // 用于演示发送操作。
		for _, elem := range []string{"a", "b", "c", "d"} {
			strChan <- elem
			fmt.Println("Sent:", elem, "[sender]")
			if elem == "c" {
				syncChan1 <- struct{}{}
				fmt.Println("Sent a sync signal. [sender]")
			}
		}
		fmt.Println("Wait 2 seconds... [sender]")
		time.Sleep(time.Second * 2)
		close(strChan)
		syncChan2 <- struct{}{}
	}()
	<-syncChan2
	<-syncChan2
}
/*
Sent: a [sender]
Sent: b [sender]
Sent: c [sender]
Sent a sync signal. [sender]
Received a sync signal and wait a second... [receiver]
Received: a [receiver]
Sent: d [sender]   strChan的容量是3，所以发送方在这1s里发送第4个值时会因strChan已满而等待，直到接收方从strChan取出第一个值
Wait 2 seconds... [sender]
Received: b [receiver]
Received: c [receiver]
Received: d [receiver]
Stopped. [receiver]
*/



