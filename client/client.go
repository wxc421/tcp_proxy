package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// 连接服务器
	conn, err := net.Dial("tcp", "127.0.0.1:9897")
	if err != nil {
		fmt.Println("Connect to TCP server failed ,err:", err)
		return
	}
	select {}

	// 读取命令行输入
	inputReader := bufio.NewReader(os.Stdin)

	// 一直读取直到遇到换行符
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println("Read from console failed,err:", err)
			return
		}

		// 读取到字符"Q"退出
		str := strings.TrimSpace(input)
		if str == "Q" {
			break
		}

		// 响应服务端信息
		_, err = conn.Write([]byte(input))
		//bytes, _ := ioutil.ReadAll(conn)
		//fmt.Println(string(bytes))
		//if err != nil {
		//	fmt.Println("Write failed,err:", err)
		//	break
		//}
		b := make([]byte, 1024)
		for {
			_, err := conn.Read(b)
			if err != nil {
				fmt.Println("read err", err)
				return
			}
			fmt.Println(string(b))
		}
	}

}
