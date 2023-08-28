package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/quic-go/quic-go"
	"go_quic_examples/util"
	"io"
	"os"
	"strings"
)

const ConnectionCloseCode quic.ApplicationErrorCode = 0x1

var quitCh = make(chan struct{})

// Quic 客户端
func main() {
	// 1. 获取tls配置
	tlsConfig, err := util.GetClientTlsConfig()
	if err != nil {
		fmt.Printf("server load tls error: %v \n", err)
		return
	}
	// 2. 连接Quic服务端
	conn, err := quic.DialAddr(context.Background(), "127.0.0.1:4433", tlsConfig, nil)
	if err != nil {
		fmt.Println("conn quic server error: " + err.Error())
		return
	}
	// 3. 为连接打开一个双向流
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("conn open stream error: " + err.Error())
		return
	}
	// 开启一个协程 读取流中服务端发来的数据
	go StreamHandler(stream)

	// 向连接的数据流中写入数据
	reader := bufio.NewReader(os.Stdin)
	for {
		line, _, _ := reader.ReadLine()
		msg := string(line)
		if strings.Contains(msg, "close") {
			// 关闭流
			stream.Close()
			break
		}
		if strings.Contains(msg, "quit") {
			// 关闭连接
			conn.CloseWithError(ConnectionCloseCode, "conn close")
			break
		}
		_, _ = stream.Write(line)
	}
	<-quitCh
}

// StreamHandler 数据流处理，负责从流中读取服务端发来的数据
func StreamHandler(stream quic.Stream) {
	buf := make([]byte, 1024)
	for {
		// 阻塞读取流中的数据
		length, err := stream.Read(buf)
		if err == nil {
			message := string(buf[:length])
			fmt.Printf("Stream[%d]: %s \n", stream.StreamID(), message)
			continue
		}
		// 错误处理
		switch e := err.(type) {
		// 自定义错误码处理
		case *quic.ApplicationError:
			// 流的连接已关闭，停止流的处理
			if e.ErrorCode == ConnectionCloseCode {
				fmt.Println("connection close~")
				quitCh <- struct{}{}
				return
			}
		// 连接或者流空闲超时
		case *quic.IdleTimeoutError:
			fmt.Println("connection timeout~")
			quitCh <- struct{}{}
			return
		default:
			// 流的发送流已关闭，关闭流的接收流,然后退出
			if err == io.EOF {
				fmt.Println("stream close~")
				quitCh <- struct{}{}
				return
			}
		}
	}
}
