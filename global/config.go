package global

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	SensitiveWords []string

	MessageQueueLen = 1024
)

func initConfig() {
	viper.SetConfigName("chatroom")
	viper.AddConfigPath(RootDir + "/config")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	SensitiveWords = viper.GetStringSlice("sensitive")
	MessageQueueLen = viper.GetInt("message-queue")

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		viper.ReadInConfig()

		SensitiveWords = viper.GetStringSlice("sensitive")
	})
}
