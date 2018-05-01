package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	fileBasedPipe()
	inMemorySyncPipe()
}

func fileBasedPipe() {
	reader, writer, err := os.Pipe() //使用系统函数来创建管道，并把它的两端封装成两个*os.File类型的值
	if err != nil {
		fmt.Printf("Error: Couldn't create the named pipe: %s\n", err)
	}
	go func() {        // 并发运行，程序要么阻塞在 Reader ，要么在 Writer 
		output := make([]byte, 100)
		n, err := reader.Read(output)  // 从管道 reader 读数据，返回读取的字节数
		if err != nil {
			fmt.Printf("Error: Couldn't read data from the named pipe: %s\n", err)
		}
		fmt.Printf("Read %d byte(s). [file-based pipe]\n", n)
	}()
	input := make([]byte, 26)
	for i := 65; i <= 90; i++ {
		input[i-65] = byte(i)
	}
	n, err := writer.Write(input)  // 往管道 writer 写数据，返回写入的字节数
	if err != nil {
		fmt.Printf("Error: Couldn't write data to the named pipe: %s\n", err)
	}
	fmt.Printf("Written %d byte(s). [file-based pipe]\n", n)
	time.Sleep(200 * time.Millisecond)
}

func inMemorySyncPipe() {
	reader, writer := io.Pipe()  // 基于内存的有原子性操作保证的管道, 简称内存管道
	go func() {
		output := make([]byte, 100)
		n, err := reader.Read(output)
		if err != nil {
			fmt.Printf("Error: Couldn't read data from the named pipe: %s\n", err)
		}
		fmt.Printf("Read %d byte(s). [in-memory pipe]\n", n)
	}()
	input := make([]byte, 26)
	for i := 65; i <= 90; i++ {
		input[i-65] = byte(i)
	}
	n, err := writer.Write(input)
	if err != nil {
		fmt.Printf("Error: Couldn't write data to the named pipe: %s\n", err)
	}
	fmt.Printf("Written %d byte(s). [in-memory pipe]\n", n)
	time.Sleep(200 * time.Millisecond)
}
