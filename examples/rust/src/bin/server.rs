use anyhow::{Result};
use s2n_quic::{Connection, Server};
use s2n_quic::stream::{BidirectionalStream};
use tokio::io::{AsyncReadExt, AsyncWriteExt};

// tls证书
const CERT_PEM: &str = include_str!("../fixtures/cert.pem");
// 证书秘钥
const KEY_PEM: &str = include_str!("../fixtures/key.pem");

///
/// 基于s2n-quic + tokio 实现的Quic服务端
///
#[tokio::main]
async fn main() -> Result<()> {
    // 1. 创建Quic服务
    let mut server = Server::builder()
        // 设置服务的证书与秘钥
        .with_tls((CERT_PEM, KEY_PEM))?
        // 设置服务端点
        .with_io("127.0.0.1:4433")?
        .start()?;
    eprintln!("Quic Server Running...");
    // 阻塞等待客户端连接
    while let Some(conn) = server.accept().await {
        // 启动一个异步任务处理连接
        tokio::spawn(connection_handler(conn));
    }
    Ok(())
}

// 连接处理函数
async fn connection_handler(mut conn: Connection) {
    eprintln!("client connection from [{:?}]", conn.remote_addr().unwrap());
    // 一个连接可能打开多个流
    loop {
        // 阻塞等待连接打开数据流
        match conn.accept_bidirectional_stream().await {
            Ok(Some(stream)) => {
                // 启动一个异步任务，处理数据流
                tokio::spawn(stream_handler(stream));
            }
            // TODO 连接错误处理
            Ok(None) => {
                eprintln!("None");
            }
            Err(e) => {
                eprintln!("{}",e);
            }
        }
    }
}

// 数据流处理函数
async fn stream_handler(mut stream: BidirectionalStream){
    let conn_addr = stream.connection().remote_addr().unwrap();
    eprintln!("connection from [{}] open stream of ID: [{}]",conn_addr,stream.id());
    let mut buf = [0u8;1024];
    loop {
        match stream.read(&mut buf).await {
            Ok(length) =>{
                // 流的发送端已关闭
                if length == 0 {
                    eprintln!("connection from [{}] of stream [{}] closed.",conn_addr,stream.id());
                    return;
                }
                let message = String::from_utf8_lossy(&mut buf[..length]).trim().to_string();
                eprintln!("[{}]-[{}]: {}",conn_addr,stream.id(),message);
                // 将数据写回流中
                _ = stream.write(&mut buf[..length]).await;
            }
            // TODO 流错误处理
            Err(e) => {
                eprintln!("{}",e)
            }
        }
    }
}