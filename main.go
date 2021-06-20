package main

import (
	"flag"
	"net"
	"os"
	"simpletcp/config"
	"simpletcp/src/devices"
	"simpletcp/src/monitor"
	"simpletcp/src/tcpserver"
	"time"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var configName = flag.String("config", "config", "配置文件名称，默认config")
var configPath = flag.String("path", "./configs", "配置文件路径，默认path")

func init() {

	// logger.SetFormatter(&logger.JSONFormatter{})

	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	logger.SetOutput(os.Stdout)

	// 设置日志级别为warn以上
	logger.SetLevel(logger.DebugLevel)
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}
	config.InitConfig(*configPath, *configName)
	go monitor.CreateMonitorServer()
	address := viper.GetString("SERVER_IP") + ":" + viper.GetString("SERVER_PORT")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}
		duration := viper.GetInt("TIMEOUT")
		if duration == 0 {
			duration = 30
		}
		logger.Infof("welcome %s join !!! ", conn.RemoteAddr())
		tcpserver.NewTcpConn(
			conn,
			tcpserver.WithCheckDuration(time.Second*time.Duration(duration)),
			tcpserver.WithOnMessage(onMessage),
			tcpserver.WithTimeout(onTimeout),
			tcpserver.WithInitMsg("please input your deviceId:"),
			tcpserver.WithChangeDeviceIDHook(onChangeDeviceID),
		)
	}
}

func onMessage(c tcpserver.TcpConn, msg []byte) {
	logger.Debugf("Device [%s] Receive : %s", c.GetDeviceId(), msg)
	switch string(msg) {
	case "q", "exit":
		logger.Infof("Device: %s 主动退出", c.GetDeviceId())
		_ = c.Close()
		devices.DeviceStore.Delete(c.GetDeviceId())
		return
	case "0100":
		logger.Debug("PING")
		break
	}
}

func onTimeout(c tcpserver.TcpConn) {
	logger.Errorf("Device: %s超时退出", c.GetDeviceId())
	c.Close()
	devices.DeviceStore.Delete(c.GetDeviceId())
}

func onChangeDeviceID(c tcpserver.TcpConn) {
	logger.Info("设置连接的 CONN DEVICE ID: ", c.GetDeviceId())
	oldConn := devices.DeviceStore.GetByDeviceID(c.GetDeviceId())
	if oldConn != nil {
		logger.Info("移除连接的 OLD CONN DEVICE ID: ", c.GetDeviceId())
		oldConn.Close()
		devices.DeviceStore.Delete(c.GetDeviceId())
	}
	devices.DeviceStore.Add(c.GetDeviceId(), c)
}
