package main

import (
	"flag"
	"fmt"
	"go_learn/tcp_proxy/tool"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var local string
var remote string

func main() {
	flag.StringVar(&local, "l", "0.0.0.0:9897", "-l=0.0.0.0:9897 指定服务监听的端口")
	flag.StringVar(&remote, "r", "127.0.0.1:3306", "-d=127.0.0.1:1789 指定后端的IP和端口")
	stopChan := make(chan os.Signal, 1) // 接收系统中断信号
	signal.Notify(stopChan, os.Interrupt)
	listener, err := net.Listen("tcp", local)
	if err != nil {
		panic(err)
	}
	log.Print("start server listen on:", local)

	go func() {
		for {
			conn, err := listener.Accept()
			log.Println("accept ...")
			if err != nil {
				fmt.Print(err)
				continue
			}
			go handle2(conn.(*net.TCPConn))
		}
	}()

	<-stopChan
	log.Println("Get Stop Command. Now Stoping...")
	if err = listener.Close(); err != nil {
		log.Print(err)
	}
}

func handle(clientConn *net.TCPConn) {
	defer func(clientConn *net.TCPConn) {
		_ = clientConn.Close()
	}(clientConn)
	serverConn, err := net.Dial("tcp", remote)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	wrapper := tool.WaitGroupWrapper{}
	wrapper.Wrap(func() {
		buf := make([]byte, 2048)
		// err == nil => 从serverConn读到EOF
		// err != nil => serverConn被server关闭导致无法read读到EOF || clientConn被client关闭导致write失败
		if _, err = io.CopyBuffer(clientConn, serverConn, buf); err != nil {
		}
		_, err := clientConn.Write([]byte{'a', 'b', 'c'})
		if err != nil {
			fmt.Println(err)
			return
		}
		_ = clientConn.Close()
		_ = serverConn.Close()
		// log.Println("serverConn -> clientConn down")
	})
	wrapper.Wrap(func() {
		buf := make([]byte, 2048)
		// err == nil => 从clientConn读到EOF
		// err != nil => clientConn被client关闭导致无法read读到EOF || serverConn被server关闭导致write失败
		if _, err = io.CopyBuffer(serverConn, clientConn, buf); err != nil {
		}
		_ = serverConn.Close()
		_ = clientConn.Close()
		// log.Println("clientConn -> serverConn down")
	})
	wrapper.Wait()
}

func handle2(clientConn *net.TCPConn) {
	defer func(clientConn *net.TCPConn) {
		_ = clientConn.Close()
	}(clientConn)
	serverConn, err := net.Dial("tcp", remote)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func(serverConn net.Conn) {
		_ = serverConn.Close()
	}(serverConn)

	wrapper := tool.WaitGroupWrapper{}
	mutex := sync.Mutex{}
	wrapper.Wrap(func() {
		_, err := copyBufferGetData(clientConn, serverConn, func(data []byte) {
			mutex.Lock()
			defer mutex.Unlock()
			fmt.Printf("[==>] Received %d bytes from %s\n", len(data), serverConn.RemoteAddr())
			hexDump(data)
			fmt.Printf("[<==] Send %d bytes to %s\n", len(data), clientConn.RemoteAddr().String())
			fmt.Println("----------------------------------------------------")
		})
		if err != nil {
			fmt.Println(err)
		}
		_ = clientConn.Close()
		_ = serverConn.Close()
		log.Println("serverConn -> clientConn down")
	})
	wrapper.Wrap(func() {
		// err == nil => 从clientConn读到EOF
		// err != nil => clientConn被client关闭导致无法read读到EOF || serverConn被server关闭导致write失败
		fmt.Printf("[==>] Received incoming connection from %s\n", serverConn.RemoteAddr())
		_, err := copyBufferGetData(serverConn, clientConn, func(data []byte) {
			mutex.Lock()
			defer mutex.Unlock()
			fmt.Printf("[==>] Received %d bytes from %s\n", len(data), clientConn.RemoteAddr())
			hexDump(data)
			fmt.Printf("[<==] Send %d bytes to %s\n", len(data), serverConn.RemoteAddr().String())
			fmt.Println("----------------------------------------------------")
		})
		if err != nil {
			fmt.Println(err)
		}
		_ = clientConn.Close()
		_ = serverConn.Close()
		log.Println("clientConn -> serverConn down")
	})
	wrapper.Wait()
}

func hexDump(data []byte) {
	step := 16
	for i := 0; i < len(data); i += step {
		var s []byte
		if i+step > len(data) {
			s = data[i:]
		} else {
			s = data[i : i+step]
		}

		s2 := make([]string, 0, step)
		texts := make([]string, 0, step)
		for _, x := range s {
			signStr := fmt.Sprintf("%02x", x) //将[]byte转成16进制
			s2 = append(s2, signStr)
			text := "."
			if x >= 0x20 && x < 0x7F {
				text = string(x)
			}
			texts = append(texts, text)
		}
		if i+step > len(data) {
			for i := 0; i < step-len(data)%16; i++ {
				s2 = append(s2, "  ")
			}
		}
		join := strings.Join(s2, " ")
		fmt.Println(fmt.Sprintf("%04x", i), join, "     ", texts)
	}
}

func copyBufferGetData(dst io.Writer, src io.Reader, fn func([]byte)) (data []byte, err error) {
	data = make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		data = make([]byte, 0, 4096)
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			data = append(data, buf[0:nr]...)
			fn(data)
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = errors.New("short buffer")
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return data, nil
}
