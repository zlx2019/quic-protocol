// @Title util.go
// @Description
// @Author Zero - 2023/8/27 19:17:31

package util

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

// GetClientTlsConfig 获取Quic客户端的tls配置
func GetClientTlsConfig() (*tls.Config, error) {
	// 读取证书文件
	file, err := os.ReadFile("cert.pem")
	if err != nil {
		return nil, err
	}
	// 创建证书池，将证书放入池中
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(file)
	// 创建tls配置
	config := &tls.Config{
		ServerName: "localhost",
		// 客户端验证服务端所需要的证书池，如果不设置则使用主机的根证书
		RootCAs: pool,
	}
	return config, nil
}

// GetServerTlsConfig 获取Quic服务端的tls配置
func GetServerTlsConfig() (*tls.Config, error) {
	// 加载tsl证书和秘钥文件
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		return nil, err
	}
	// tls的key信息，写入到session.log文件中，用于抓包
	logFile, err := os.OpenFile("session.log", os.O_CREATE|os.O_WRONLY, 0666)
	// tls配置
	config := &tls.Config{
		ServerName:   "localhost",             // 服务名
		Certificates: []tls.Certificate{cert}, // tls的证书和秘钥
		KeyLogWriter: logFile,
	}
	return config, nil
}
