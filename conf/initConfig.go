package conf

import (
	"JetIotDeviceSimulator/Log"
	"encoding/json"
	_ "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config_t struct {
	ClientReplyTopic         string `json:"client_reply_topic"`
	ClientRegisterReplyTopic string `json:"client_register_reply_topic"`
	ClientRegisterTopic      string `json:"client_register_topic"`
	ClientPublishTopic       string `json:"client_publish_topic"`
	ServerWill               string `json:"server_will"`
	MqttServer               string `json:"mqtt_server"`
	MqttPort                 string `json:"mqtt_port"`
	MqttClientID             string `json:"mqtt_client_id"`
	MqttUserName             string `json:"mqtt_user_name"`
	MqttPassword             string `json:"mqtt_password"`
}

var (
	Default Config_t
)

func InitConfig(path string) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		Log.E()("读取文件失败", err.Error())
		return
	}
	Log.I()("文件读取成功")

	Default = Config_t{}
	err = json.Unmarshal(file, &Default)
	if err != nil {
		Log.E()("解析配置文件失败", err.Error())
	}
	Log.I()("解析配置文件成功")
}
