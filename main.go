package main

import (
	"JetIotDeviceSimulator/Log"
	"JetIotDeviceSimulator/conf"
	"JetIotDeviceSimulator/model"
	"JetIotDeviceSimulator/model/event"
	"JetIotDeviceSimulator/model/response"
	"JetIotDeviceSimulator/mqtt"
	"encoding/json"
	mq "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
	"time"
)

var configPath = pflag.StringP("config", "c", "conf/defaultConfig.json", "the config json file path")

func main() {
	go func() {
		conf.InitConfig(*configPath)
		mqtt.InitMqttClient()
		mqtt.Subscribe(conf.Default.ServerWill, tryReConnect)
		mqtt.Subscribe(conf.Default.ClientRegisterReplyTopic, registerReplyCallBack)
		mqtt.RegisterEventHandle(event.EVENT_THING_DEVICE_ONLIONE, "onlineCallBack", onlineCallBack)
		mqtt.RegisterEventHandle(event.EVENT_COMPONENT_CHANGE_VALUE, "setValue", setValue)
		mqtt.EventListenStart()
		Connect()
		go func() {
			temp := 1
			for true {
				temp++
				if temp == 100 {
					temp = 1
				}
				ReportTemp(temp)
				time.Sleep(time.Duration(2) * time.Second)
			}
		}()
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func ReportTemp(value int) {
	ov := model.ThingComponentValueOV{
		Id:            conf.Default.MqttClientID,
		ComponentName: "temp",
		Value:         value,
	}
	ov.EventId = event.EVENT_COMPONENT_CHANGE_VALUE
	marshal, _ := json.Marshal(ov)
	mqtt.Publish(conf.Default.ClientPublishTopic, string(marshal))
}

func tryReConnect(client mq.Client, message mq.Message) {
	if string(message.Payload()) == "lose_connect" {
		Log.E()(string("失去与服务器链接"))
	} else if string(message.Payload()) == "server_start" {
		Log.I()(string("连接到服务器"))
	}
}

func setValue(client mq.Client, message mq.Message) {

}

func registerDevice() {
	//注册设备
	thing := model.Thing{
		Name: "设备模拟器",
		Id:   conf.Default.MqttClientID,
	}
	component := model.Component{
		Name:  "temp",
		Type:  "int",
		Value: 22,
	}
	componentList := make(map[string]*model.Component)
	componentList["temp"] = &component
	thing.Components = componentList

	marshal, _ := json.Marshal(thing)
	mqtt.Publish(conf.Default.ClientRegisterTopic, string(marshal))
	Connect()
}

func Connect() {
	ov := model.ThingInitOV{
		Id: conf.Default.MqttClientID,
	}
	ov.EventId = event.EVENT_THING_DEVICE_ONLIONE
	marshal, _ := json.Marshal(ov)
	mqtt.Publish(conf.Default.ClientPublishTopic, string(marshal))
}

func onlineCallBack(client mq.Client, message mq.Message) {
	res := response.Responses_t{}
	err := mqtt.ShouldBindJSON(message, &res)
	if err != nil {
		Log.E()(err)
	}
	if res.Success {
		Log.I()("上线成功")
	} else {
		marshal, _ := json.Marshal(res)
		Log.I()("上线失败未注册，开始注册", string(marshal))
		registerDevice()
	}
}

func registerReplyCallBack(client mq.Client, message mq.Message) {
	res := response.Responses_t{}
	err := mqtt.ShouldBindJSON(message, &res)
	if err != nil {
		Log.E()(err)
	}
	Log.I()(res.Msg)
}
