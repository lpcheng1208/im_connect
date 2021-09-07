package util

import (
	"encoding/json"
	"im_connect/pkg/logger"
	"im_connect/pkg/pb"

	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"

	"go.uber.org/zap"
)

func JsonMarshal(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		logger.Logger.Error("json序列化：", zap.Error(err))
	}
	return Bytes2str(bytes)
}

func FormatMessage(messageType pb.ImMessageType, messageContent []byte) string {
	if messageType == pb.ImMessageType_IM_MT_UNKNOWN {
		logger.Logger.Error("error message type")
		return "error message type"
	}
	var (
		msg proto.Message
		err error
	)
	switch messageType {
	case pb.ImMessageType_IM_MT_TEXT:
		msg = &pb.ImText{}
		err = proto.Unmarshal(messageContent, msg)
	case pb.ImMessageType_IM_MT_FACE:
		msg = &pb.ImFace{}
		err = proto.Unmarshal(messageContent, msg)
	case pb.ImMessageType_IM_MT_VOICE:
		msg = &pb.ImVoice{}
		err = proto.Unmarshal(messageContent, msg)
	case pb.ImMessageType_IM_MT_IMAGE:
		msg = &pb.ImImage{}
		err = proto.Unmarshal(messageContent, msg)
	case pb.ImMessageType_IM_MT_FILE:
		msg = &pb.ImFile{}
		err = proto.Unmarshal(messageContent, msg)
	case pb.ImMessageType_IM_MT_LOCATION:
		msg = &pb.ImLocation{}
		err = proto.Unmarshal(messageContent, msg)
	case pb.ImMessageType_IM_MT_COMMAND:
		msg = &pb.ImCommand{}
		err = proto.Unmarshal(messageContent, msg)
	case pb.ImMessageType_IM_MT_CUSTOM:
		msg = &pb.ImCustom{}
		err = proto.Unmarshal(messageContent, msg)
	}

	bytes, err := jsoniter.Marshal(msg)
	if err != nil {
		logger.Sugar.Error(err)
		return ""
	}
	return Bytes2str(bytes)
}
