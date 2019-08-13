package protocol

import (
	"reflect"
	. "study/basic_server/gohipernetFake"
)

const (
	MAX_USER_ID_BYTE_LENGTH      = 16
	MAX_USER_PW_BYTE_LENGTH      = 16
	MAX_CHAT_MESSAGE_BYTE_LENGTH = 126
)

var _ClientSessionHeaderSize int16
var _ServerSessionHeaderSize int16

type Packet struct {
	UserSessionIndex    int32
	UserSessionUniqueId uint64
	Id                  int16
	DataSize            int16
	Data                []byte
}

type Header struct {
	TotalSize  int16
	ID         int16
	PacketType int8 // 비트 필드로 데이터 설정 0이면 Normal 1번 비트 On(압축) , 2번 비트 On(암호화)
}

func Init_packet() {
	_ClientSessionHeaderSize = protocolInitHeaderSize()
	_ServerSessionHeaderSize = protocolInitHeaderSize()
}

func protocolInitHeaderSize() int16 {
	var packetHeader Header
	headerSize := Sizeof(reflect.TypeOf(packetHeader))
	return (int16)(headerSize)
}

/// [방 입장]
type RoomEnterReqPacket struct {
	RoomNumber int32
}

func (request *RoomEnterReqPacket) Decoding(bodyData []byte) bool {
	if len(bodyData) != (4) {
		return false
	}

	reader := MakeReader(bodyData, true)
	request.RoomNumber, _ = reader.ReadS32()

	return true
}

type RoomEnterResPacket struct {
	Result           int16
	RoomNumber       int32
	RoomUserUniqueId uint64
}

func (response RoomEnterResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _ClientSessionHeaderSize + 2 + 4 + 8
	sendBuf := make([]byte, totalSize)

	writer := MakeWriter(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_ROOM_ENTER_RES, 0)
	writer.WriteS16(response.Result)
	writer.WriteS32(response.RoomNumber)
	writer.WriteU64(response.RoomUserUniqueId)
	return sendBuf, totalSize
}

func EncodingPacketHeader(writer *RawPacketData, totalSize int16, pktId int16, packetType int8) {
	writer.WriteS16(totalSize)
	writer.WriteS16(pktId)
	writer.WriteS8(packetType)
}
