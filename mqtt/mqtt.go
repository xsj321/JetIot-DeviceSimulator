package mqtt

import (
	"JetIotDeviceSimulator/Log"
	"JetIotDeviceSimulator/conf"
	"JetIotDeviceSimulator/model"
	"encoding/json"
	mq "github.com/eclipse/paho.mqtt.golang"
)

var Client mq.Client

type handleFunc struct {
	funcName string
	handle   mq.MessageHandler
}

var EventHandleList map[int]handleFunc

func willHandle() string {
	ov := model.ThingInitOV{}
	ov.Id = conf.Default.MqttClientID
	ov.EventId = 4
	marshal, _ := json.Marshal(&ov)
	return string(marshal)
}

func InitMqttClient() {
	EventHandleList = make(map[int]handleFunc, 0)
	opts := mq.NewClientOptions()
	opts.AddBroker("tcp://"+conf.Default.MqttServer+":"+conf.Default.MqttPort).
		SetClientID(conf.Default.MqttClientID).
		SetUsername(conf.Default.MqttUserName).
		SetPassword(conf.Default.MqttPassword).
		SetWill(conf.Default.ClientPublishTopic, willHandle(), 2, false)

	Client = mq.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		Log.E()("初始化Mqtt客户端失败" + token.Error().Error())
		return
	}
	Log.I()("初始化Mqtt客户端完成")
}

func Publish(topic string, msg string) {
	Log.D()("topic：" + topic + " msg：" + msg)
	token := Client.Publish(topic, 1, false, msg)
	if token.Error() != nil {
		Log.E()("发布消息失败" + token.Error().Error())
		return
	}
	if !token.Wait() {
		Log.E()("发布发送等待")
		return
	}
}

func Subscribe(topic string, callback mq.MessageHandler) {
	if token := Client.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		Log.E()("订阅失败" + token.Error().Error())
	}
}

func Unsubscribe(topic string) {
	if token := Client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
		Log.E()("取消订阅失败" + token.Error().Error())
	}
}

func Replay(topic string, replay interface{}) {
	marshal, err := json.Marshal(replay)
	if err != nil {
		Log.E()("回复失败" + err.Error())
		return
	}
	Publish(topic, string(marshal))
}

func ReplayToDevice(deviceId string, replay interface{}) {
	marshal, err := json.Marshal(replay)
	if err != nil {
		Log.E()("回复失败" + err.Error())
		return
	}
	Publish("thing/entity/"+deviceId+"/todevice", string(marshal))
}

/**
 * @Description: 注册订阅事件回调函数
 * @param eventId 事件ID
 * @param eventName 事件名称
 * @param handle 回调函数
 */
func RegisterEventHandle(eventId int, eventName string, handle mq.MessageHandler) {
	Log.I()("注册事件：", eventName, "        \t--> EVENT_ID：", eventId)
	handleFunc := handleFunc{
		funcName: eventName,
		handle:   handle,
	}
	EventHandleList[eventId] = handleFunc
}

func EventListenStart() {
	Log.I()("初始化事件系统")
	Subscribe(conf.Default.ClientReplyTopic, DealEventHandle)
}

func DealEventHandle(client mq.Client, message mq.Message) {
	Log.D()(" json:", string(message.Payload()))
	Log.D()("出现事件")
	payload := struct {
		EventId int `json:"event_id"`
	}{}
	err := ShouldBindJSON(message, &payload)
	if err != nil {
		Log.E()("解析事件ID错误")
	}
	Log.D()("Event ID :", payload.EventId, " json:", string(message.Payload()))
	handleFunc, exist := EventHandleList[payload.EventId]
	if !exist {
		Log.E()("EevntId：", payload.EventId, " 不存在")
	}
	Log.D()("调用方法: ", handleFunc.funcName)
	handleFunc.handle(client, message)
}

/**
 * @Description: 解析来着设备传递的Json字符串到对象
 * @param message 数据对象
 * @param model 解析目标对象的指针
 * @return error
 */
func ShouldBindJSON(message mq.Message, model interface{}) error {
	payload := message.Payload()
	err := json.Unmarshal(payload, model)
	return err
}
