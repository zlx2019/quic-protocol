package main

import (
	"context"
	"fmt"
	"github.com/quic-go/quic-go"
	"go_quic_examples/util"
	"io"
	"time"
)

const ConnectionCloseCode quic.ApplicationErrorCode = 0x1

// Quic 服务端
func main() {
	// 1. 获取tls配置
	tlsConfig, err := util.GetServerTlsConfig()
	if err != nil {
		fmt.Printf("server load tls error: %v \n", err)
		return
	}
	// 2.设置quic服务配置
	config := &quic.Config{
		// 连接空闲超时时间
		// 超时后连接和连接下的流的操作返回IdleTimeoutError错误
		MaxIdleTimeout: time.Second * 9999,
	}

	// 3.创建quic服务
	server, err := quic.ListenAddr("127.0.0.1:4433", tlsConfig, config)
	if err != nil {
		fmt.Println("listener quic server error: " + err.Error())
		return
	}

	fmt.Println("Quic Server Running Successful...")

	// 4.阻塞等待客户端连接
	for {
		conn, err := server.Accept(context.Background())
		if err != nil {
			fmt.Println("server accept connection error: " + err.Error())
			continue
		}
		// 5. 开启协程处理客户端连接
		go ConnectionHandler(conn)
	}
}

// ConnectionHandler 连接处理
// 负责处理连接中打开的流
func ConnectionHandler(conn quic.Connection) {
	fmt.Printf("[%s] client connectd... \n", conn.RemoteAddr())
	// 一个连接可以打开多个流，需要循环处理
	for {
		// 阻塞等待连接打开一个双向数据流，当客户端打开流并且发送数据时才可以等待到
		// 如果客户端连接已关闭或超时，则会返回错误
		stream, err := conn.AcceptStream(context.Background())
		if err == nil {
			// 开启协程处理数据流
			go StreamHandler(conn, stream)
			continue
		}
		// 连接的错误处理
		switch e := err.(type) {
		// 连接已关闭错误
		case *quic.ApplicationError:
			if e.ErrorCode == ConnectionCloseCode {
				fmt.Printf("[%s] connection closed~ \n", conn.RemoteAddr())
				return
			}
		// 连接空闲超时
		case *quic.IdleTimeoutError:
			fmt.Printf("[%s] connection timeout~ \n", conn.RemoteAddr())
			return
		}

	}
}

// StreamHandler 连接的数据流处理
// 负责读取流中的数据
func StreamHandler(conn quic.Connection, stream quic.Stream) {
	fmt.Printf("[%s] open stream of ID: [%d]... \n", conn.RemoteAddr(), stream.StreamID())
	// 创建一个缓冲区，用于读取流中的数据
	buf := make([]byte, 1024)
	for {
		// 阻塞等待读取流中的数据
		length, err := stream.Read(buf)
		if err == nil {
			message := string(buf[:length])
			fmt.Printf("[%s]-[%d]: %s \n", conn.RemoteAddr(), stream.StreamID(), message)
			// 写回流中
			_, _ = stream.Write(buf[:length])
			continue
		}
		// 流的错误处理
		switch e := err.(type) {
		// 自定义错误码处理
		case *quic.ApplicationError:
			// 流的连接已关闭，停止流的处理
			if e.ErrorCode == ConnectionCloseCode {
				fmt.Printf("[%s]-[%d] stream shutdown for connection close~ \n", conn.RemoteAddr(), stream.StreamID())
				return
			}
		// 连接或者流空闲超时
		case *quic.IdleTimeoutError:
			fmt.Printf("[%s]-[%d] stream timeout~ \n", conn.RemoteAddr(), stream.StreamID())
			return
		default:
			// 流的发送流已关闭，关闭流的接收流,然后退出
			if err == io.EOF {
				_ = stream.Close() // 关闭流的接收端
				fmt.Printf("[%s]-[%d] stream close~ \n", conn.RemoteAddr(), stream.StreamID())
				return
			}
		}
	}
}
