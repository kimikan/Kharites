package main

import (
	"Kharites/ioservice"
	"bufio"
	"fmt"
	"os"
	"testing"
	"time"
)

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	//defer reader.Close()
	strBytes, _, err := reader.ReadLine()

	if err == nil {
		return string(strBytes)
	}

	return ""
}

func TestDevRead(test *testing.T) {
	dev := "/home/kan/MyProjects/Diskless/root/large.img"
	dev = readLine()
	fmt.Println(dev)

	const blockSize = 64 * 1024
	bytes := make([]byte, blockSize)
	var offset int64

	start := time.Now()
	for true {
		offset += blockSize
		result := ioservice.ReadBytesFromFile(dev, bytes, offset)
		if result <= 0 {
			break
		}
	}
	dis := time.Now().Sub(start).Seconds()
	d := (float64(offset)) / 1024 / 1024
	fmt.Println("Read size(mb): ", d, "  Seconds: ", dis, " mb/s: ", d/float64(dis))
}
