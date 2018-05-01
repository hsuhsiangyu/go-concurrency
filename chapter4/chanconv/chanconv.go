package main

import "fmt"

func main() {
	var ok bool
	ch := make(chan int, 1)
	_, ok = interface{}(ch).(<-chan int)
	fmt.Println("chan int => <-chan int:", ok)
	_, ok = interface{}(ch).(chan<- int)
	fmt.Println("chan int => chan<- int:", ok)

	sch := make(chan<- int, 1)
	_, ok = interface{}(sch).(chan int)
	fmt.Println("chan<- int => chan int:", ok)

	rch := make(<-chan int, 1)
	_, ok = interface{}(rch).(chan int)
	fmt.Println("<-chan int => chan int:", ok)
}
/*
chan int => <-chan int: false   类型都不一样
chan int => chan<- int: false
chan<- int => chan int: false
<-chan int => chan int: false
*/

