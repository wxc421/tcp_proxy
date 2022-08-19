package main

import (
	"bufio"
	"fmt"
	"net"
)

/*
TCP服务端程序的处理流程：

监听端口
接收客户端请求建立链接
创建goroutine处理链接。
*/
//创造处理链接的函数
func process(conn net.Conn) {
	//2.最后要关闭这个链接
	defer conn.Close()
	//1.根据链接进行数据的接收和发送操作
	//1.1根据这个tcp链接conn,创建一个从链接conn读的对象
	reader := bufio.NewReader(conn)
	//1.
	var buf [128]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		fmt.Printf("read from conn  failed,err:%d", err)
	}
	ret := buf[:n]
	fmt.Printf("从链接读取的数据是：%v", string(ret))

	conn.Write([]byte("亲亲，收到了"))

}
func main() {
	//1.启动监听  先放上耳朵
	listener, err := net.Listen("tcp", "127.0.0.1:1789")
	if err != nil {
		fmt.Printf("listen failed,err:%v", err)
		return
	}
	//2.等待客户建立连接  //for不断的建立链接处理链接
	for {
		conn, err := listener.Accept() //如果err为空，代表拿到了链接
		if err != nil {
			fmt.Printf("conn accept failed,err:%v", err)
			continue
		}
		//3.启动goroutine去处理链接
		go process(conn)

	}

}
