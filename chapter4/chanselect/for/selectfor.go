package main

import "fmt"

func main() {
	intChan := make(chan int, 10)
	for i := 0; i < 10; i++ {
		intChan <- i
	}
	close(intChan)
	syncChan := make(chan struct{}, 1)
	go func() {
	Loop:
		for {
			select {
			case e, ok := <-intChan:
				if !ok {
					fmt.Println("End.")
					break Loop // break Loop会中断紧贴于该标签之下的for语句. 如果不带标签，只能中断所在select语句的执行
				}
				fmt.Printf("Received: %v\n", e)
			}
		}
		syncChan <- struct{}{}
	}()
	<-syncChan
}
