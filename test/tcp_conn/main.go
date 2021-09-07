package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"im_connect/pkg/pb"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	util2 "im_connect/pkg/gn/test/util"
)

var (
	RedisCli *redis.Client
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	util2.MaxLen = 2048
	//InitRedis("127.0.0.1:6379", "")

	client := TcpClient{}
	//log.Println("input UserId,DeviceId,SyncSequence")
	//log.Scanf("%d %d %d", &client.UserId, &client.DeviceId, &client.Seq)
	client.UserId = fmt.Sprintf("hello%d", 1)
	client.DeviceId = "hello:DeviceId"
	client.Seq = 0
	client.Start()
	time.Sleep(time.Millisecond * 500)


	//defer RedisCli.Close()

	select {}
}

func Json(i interface{}) string {
	bytes, _ := jsoniter.Marshal(i)
	return string(bytes)
}

type TcpClient struct {
	UserId   string
	DeviceId string
	Seq      int64
	codec    *util2.Codec
}

func (c *TcpClient) Output(pt pb.IMPackageType, requestId int64, message proto.Message) {
	var input = pb.ImInput{
		Type:      pt,
		RequestId: requestId,
	}
	if message != nil {
		bytes, err := proto.Marshal(message)
		if err != nil {
			log.Println(err)
			return
		}
		input.Data = bytes
	}

	inputByf, err := proto.Marshal(&input)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = c.codec.Conn.Write(util2.Encode(inputByf))
	if err != nil {
		log.Println(err)
	}
}

func (c *TcpClient) Start() {
	connect, err := net.Dial("tcp", "127.0.0.1:9000")

	if err != nil {
		log.Println(err)
		return
	}

	c.codec = util2.NewCodec(connect)

	c.SignIn()
	//time.Sleep(time.Second)
	//c.SyncTrigger()
	//c.SubscribeRoom()
	//c.SendMsg()

	go c.Heartbeat()
	//go c.Receive()
}

func (c *TcpClient) SignIn() {
	signIn := pb.ImSignInInput{
		UserId:   c.UserId,
		DeviceId: c.DeviceId,
		Token:    "1",
	}
	c.Output(pb.IMPackageType_IM_SIGN_IN, time.Now().UnixNano(), &signIn)
}

func InitRedis(addr, password string) {
	log.Println("init redis")
	RedisCli = redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		Password: password,
	})

	_, err := RedisCli.Ping().Result()
	if err != nil {
		panic(err)
	}

	log.Println("init redis ok")
}

func (c *TcpClient) SendMsg() {
	sender := &pb.ImSender{
		SenderType: pb.ImSenderType_IM_ST_USER,
		SenderId:   "hello1",
		AvatarUrl:  "1",
		Nickname:   "2",
		Extra:      "3",
	}
	text := &pb.ImText{Text: "hello"}
	MessageContent, _ := proto.Marshal(text)

	message := &pb.ImMessage{
		Sender:         sender,
		ReceiverType:   pb.ImReceiverType_IM_RT_USER,
		ReceiverId:     "helloUserId:1",
		ToUserIds:      nil,
		MessageType:    pb.ImMessageType_IM_MT_TEXT,
		MessageContent: MessageContent,
		SendTime:       time.Now().UnixNano(),
	}
	//bs, _ := proto.Marshal(message)
	//_, err := RedisCli.Publish("push_message_forward_topic", string(bs)).Result()
	//if err != nil {
	//	log.Println(err.Error())
	//	return
	//}
	c.Output(pb.IMPackageType_IM_MESSAGE, time.Now().UnixNano(), message)
}

func (c *TcpClient) Heartbeat() {
	ticker := time.NewTicker(time.Second * 30)
	for range ticker.C {
		c.Output(pb.IMPackageType_IM_HEARTBEAT, time.Now().UnixNano(), nil)
		//c.SendMsg()
	}
}

func (c *TcpClient) Receive() {
	for {
		_, err := c.codec.Read()
		if err != nil {
			log.Println(err)
			return
		}

		for {
			bytes, ok, err := c.codec.Decode()
			if err != nil {
				log.Println(err)
				return
			}

			if ok {
				c.HandlePackage(bytes)
				continue
			}
			break
		}
	}
}

func (c *TcpClient) HandlePackage(bytes []byte) {
	var output pb.ImOutput
	err := proto.Unmarshal(bytes, &output)
	if err != nil {
		log.Println(err)
		return
	}
	RequestId := output.RequestId
	switch output.Type {
	case pb.IMPackageType_IM_SIGN_IN:

		data := output.Data
		signInResp := pb.ImSignInOutput{}
		err := proto.Unmarshal(data, &signInResp)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("登录响应, RequestId: ", RequestId, " signInResp: ", signInResp.Success, " Msg: ", output.Message)
	case pb.IMPackageType_IM_HEARTBEAT:
		log.Println("心跳响应, RequestId: ", RequestId)
	case pb.IMPackageType_IM_MESSAGE:
		var ImMessage pb.ImMessage
		err := proto.Unmarshal(output.Data, &ImMessage)
		if err != nil {
			log.Println(err.Error())
			return
		}

		ReceiverId := ImMessage.ReceiverId
		MessageContent := ImMessage.MessageContent
		var msgText pb.ImText
		err = proto.Unmarshal(MessageContent, &msgText)
		if err != nil {
			log.Println(err.Error())
			return
		}
		log.Println("收到普通消息: ", " sendId: ", ImMessage.Sender.SenderId, " ReceiverId: ", ReceiverId, " msgText: ", msgText.Text, " requestId: ", RequestId)
	default:
		log.Println("switch other")
	}
}
