package main

import (
	"fmt"
	"time"
)

var mapChan = make(chan map[string]int, 1) //mapChan的元素类型属于引用类型。因此接收方对元素值的副本的修改会影响到发送方持有的赋值

func main() {
	syncChan := make(chan struct{}, 2)
	go func() { // 用于演示接收操作。
		for {
			if elem, ok := <-mapChan; ok {
				elem["count"]++
			} else {
				break
			}
		}
		fmt.Println("Stopped. [receiver]")
		syncChan <- struct{}{}
	}()
	go func() { // 用于演示发送操作。
		countMap := make(map[string]int)
		for i := 0; i < 5; i++ {
			mapChan <- countMap
			time.Sleep(time.Millisecond)
			fmt.Printf("The count map: %v. [sender]\n", countMap)
		}
		close(mapChan)
		syncChan <- struct{}{}
	}()
	<-syncChan
	<-syncChan
}
/*
The count map: map[count:1]. [sender]
The count map: map[count:2]. [sender]
The count map: map[count:3]. [sender]
The count map: map[count:4]. [sender]
The count map: map[count:5]. [sender]
Stopped. [receiver]
*/

