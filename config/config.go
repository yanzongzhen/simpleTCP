package config

import (
	"github.com/spf13/viper"
	"sync"
)

var once = new(sync.Once)

func InitConfig(path, configName string) {
	once.Do(func() {
		if len(configName) == 0 {
			configName = "config"
		}
		viper.SetConfigName(configName)
		if len(path) != 0 {
			viper.AddConfigPath(path)
		}
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("./config")
		err := viper.ReadInConfig()
		if err != nil {
			panic(err)
		}
		viper.WatchConfig()
		// debug := viper.GetBool("logger.debug")
		// level := viper.GetString("logger.level")
		// logger.InitNormalLogConfig(getLevel(level),debug)
	})
}

// func getLevel(level string) logger.LogLevel {
// 	switch level {
// 	case "debug":
// 		return logger.DEBUG
// 	case "info":
// 		return logger.INFO
// 	case "warn":
// 		return logger.WARN
// 	case "error":
// 		return logger.ERROR
// 	case "fatal":
// 		return logger.FATAL
// 	default:
// 		return logger.INFO
// 	}
// }
