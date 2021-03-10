package net

import (
	"encoding/binary"
	"fmt"

	"github.com/faymajun/gonano/core/packet"
	"github.com/faymajun/gonano/core/route"
	"github.com/faymajun/gonano/message"

	"github.com/golang/protobuf/proto"
)

const (
	MsgHeadSize = 4 //消息长度
	MsgIdSize   = 2 //消息Id
)

var OrdinaryCodecFactory = func() Codec { return NewOrdinaryCoder() }
var OrdinaryCodecErrorNotClose = func() Codec { return NewOrdinaryCoderErrorNotClose() }

//var BattleUdpCodecFactory = func() Codec { return NewBattleCoder() }
//var RobotCodecFactory = func() Codec { return NewRobotCoder() }

type Codec interface {
	Encode(msgid message.MSGID, payload []byte) []byte
	Decode(data []byte) (packet.RecvMessage, error)
}

type OrdinaryCoder struct {
	decErrNotClose bool
}

func NewOrdinaryCoder() *OrdinaryCoder {
	return &OrdinaryCoder{}
}

func NewOrdinaryCoderErrorNotClose() *OrdinaryCoder {
	return &OrdinaryCoder{decErrNotClose: true}
}

func (OrdinaryCoder) Encode(msgid message.MSGID, payload []byte) []byte {
	data := make([]byte, MsgHeadSize+MsgIdSize+len(payload))
	binary.BigEndian.PutUint32(data, uint32(MsgIdSize+len(payload)))
	binary.BigEndian.PutUint16(data[MsgHeadSize:], uint16(msgid))
	copy(data[MsgHeadSize+MsgIdSize:], payload)
	return data
}

func (o OrdinaryCoder) Decode(data []byte) (packet.RecvMessage, error) {
	msgid := binary.BigEndian.Uint16(data[0:MsgIdSize])
	handler, err := route.FindHandler(message.MSGID(msgid))
	if err != nil {
		return packet.EmptyRecvPack, err
	}

	size := len(data)
	pbMsg := handler.Instance()
	pbbuf := proto.NewBuffer(data[MsgIdSize:])
	if err := pbbuf.Unmarshal(pbMsg); err != nil {
		if o.decErrNotClose {
			return packet.EmptyRecvPack, nil
		} else {
			return packet.EmptyRecvPack, fmt.Errorf("解码错误: MsgId=%d, Length=%v Error=%v", msgid, size-MsgIdSize, err)
		}

	}

	return packet.RecvMessage{Handler: handler, Payload: pbMsg}, nil
}
