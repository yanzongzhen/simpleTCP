package monitor

import (
	logger "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"simpletcp/src/devices"
	"simpletcp/src/tcpserver"
)

type cmd struct {
	DeviceID string `json:"deviceId"`
	CMD      string `json:"cmd"`
}

const (
	OPENLOCK  = "0108"
	CLOSELOCK = "0107"
	QUERYLOCK = "0109"
)

func CreateMonitorServer() {
	e := gin.New()
	e.Use(gin.Recovery())
	e.GET("/deviceList", func(context *gin.Context) {
		val := devices.DeviceStore.List()
		context.JSON(200, gin.H{
			"code":    200,
			"message": "success",
			"data":    val,
		})
	})
	e.POST("/device/send", func(context *gin.Context) {
		var p cmd
		if err := context.Bind(&p); err != nil {
			context.JSON(400, gin.H{
				"code":    400,
				"message": "参数异常",
			})
			return
		}
		conn := devices.DeviceStore.GetByDeviceID(p.DeviceID)
		if conn == nil {
			context.JSON(400, gin.H{
				"code":    400,
				"message": "未找到设备",
			})
			return
		}
		switch p.CMD {
		case OPENLOCK:
			logger.Debugf("OPEN %s LOCK", conn.GetDeviceId())
			if conn.GetDeviceStatus() == tcpserver.Open {
				context.JSON(400, gin.H{
					"code":    400,
					"message": "设备已处在开锁状态无需处理",
				})
				return
			}
			break
		case CLOSELOCK:
			logger.Debugf("CLOSE %s LOCK", conn.GetDeviceId())
			if conn.GetDeviceStatus() == tcpserver.Close {
				context.JSON(400, gin.H{
					"code":    400,
					"message": "设备已处在关锁状态无需处理",
				})
				return
			}
			break
		case QUERYLOCK:
			logger.Debugf("QUERY %s LOCK", conn.GetDeviceId())
			break
		default:
			context.JSON(400, gin.H{
				"code":    400,
				"message": "暂不支持的指令",
			})
			return
		}
		// data, _ := hex.DecodeString(p.CMD)
		_, _ = conn.Write([]byte(p.CMD))
		if p.CMD == "0109" {
			context.JSON(200, gin.H{
				"code":    200,
				"message": "success",
				"data":    conn.GetProperty(),
			})
		} else {
			context.JSON(200, gin.H{
				"code":    200,
				"message": "success",
			})
		}
	})
	e.GET("/device/send", func(context *gin.Context) {
		p := cmd{}

		id := context.Query("id")
		cmd := context.Query("cmd")

		p.DeviceID = id
		p.CMD = cmd

		conn := devices.DeviceStore.GetByDeviceID(p.DeviceID)
		if conn == nil {
			context.JSON(400, gin.H{
				"code":    400,
				"message": "未找到设备",
			})
			return
		}
		switch p.CMD {
		case OPENLOCK:
			logger.Debugf("OPEN %s LOCK", conn.GetDeviceId())
			if conn.GetDeviceStatus() == tcpserver.Open {
				context.JSON(400, gin.H{
					"code":    400,
					"message": "设备已处在开锁状态无需处理",
				})
				return
			}
			break
		case CLOSELOCK:
			logger.Debugf("CLOSE %s LOCK", conn.GetDeviceId())
			if conn.GetDeviceStatus() == tcpserver.Close {
				context.JSON(400, gin.H{
					"code":    400,
					"message": "设备已处在关锁状态无需处理",
				})
				return
			}
			break
		case QUERYLOCK:
			logger.Debugf("QUERY %s LOCK", conn.GetDeviceId())
			break
		default:
			context.JSON(400, gin.H{
				"code":    400,
				"message": "暂不支持的指令",
			})
			return
		}
		// data, _ := hex.DecodeString(p.CMD)
		_, _ = conn.Write([]byte(p.CMD))
		if p.CMD == "0109" {
			context.JSON(200, gin.H{
				"code":    200,
				"message": "success",
				"data":    conn.GetProperty(),
			})
		} else {
			context.JSON(200, gin.H{
				"code":    200,
				"message": "success",
			})
		}
	})
	go e.Run(":" + viper.GetString("MONITOR_PORT"))
}
