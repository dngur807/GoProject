package protocol

import (
	. "GoStudy/basic_server/gohipernetFake"
	"encoding/binary"
	"reflect"
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

// 채널에 있는 유저들에게 새 유저의 정보를 알려준다.
type RoomNewUserNtfPacket struct {
	User []byte // RoomUserInfoPktData
}

type RoomUserListNtfPacket struct {
	UserCount int8
	UserList  []byte
}

func (notify RoomNewUserNtfPacket) EncodingPacket(userInfoSize int16) ([]byte, int16) {
	totalSize := _ClientSessionHeaderSize + userInfoSize
	sendBuf := make([]byte, totalSize)
	writer := MakeWriter(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_ROOM_LEAVE_RES, 0)
	return sendBuf, totalSize

}

func (notify RoomUserListNtfPacket) EncodingPacket(userInfoListSize int16) ([]byte, int16) {
	bodySize := 1 + userInfoListSize
	totalSize := _ClientSessionHeaderSize + bodySize
	sendBuf := make([]byte, totalSize)

	writer := MakeWriter(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_ROOM_USER_LIST_NTF, 0)
	writer.WriteS8(notify.UserCount)
	writer.WriteBytes(notify.UserList)
	return sendBuf, totalSize
}

// [[로그인]]  PACKET_ID_LOGIN_REQ
type LoginReqPacket struct {
	UserID []byte
	PassWD []byte
}

type LoginResPacket struct {
	Result int16
}

func (loginReq *LoginReqPacket) Decoding(bodyData []byte) bool {
	bodySize := MAX_USER_ID_BYTE_LENGTH + MAX_USER_PW_BYTE_LENGTH

	if len(bodyData) != bodySize {
		return false
	}

	reader := MakeWriter(bodyData, true)
	loginReq.UserID = reader.ReadBytes(MAX_USER_ID_BYTE_LENGTH)
	loginReq.PassWD = reader.ReadBytes(MAX_USER_PW_BYTE_LENGTH)
	return true
}

func (loginRes LoginResPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _ClientSessionHeaderSize + 2
	sendBuf := make([]byte, totalSize)

	writer := MakeWriter(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_LOGIN_RES, 0)
	writer.WriteS16(loginRes.Result)
	return sendBuf, totalSize
}

func (response ErrorNtfPacket) EncodingPacket(errorCode int16) ([]byte, int16) {
	totalSize := _ClientSessionHeaderSize + 2
	sendBuf := make([]byte, totalSize)

	writer := MakeWriter(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_ERROR_NTF, 0)
	writer.WriteS16(errorCode)
	return sendBuf, totalSize
}

type ErrorNtfPacket struct {
	ErrorCode int16
}

func NotifyErrorPacket(sessionIndex int32, sessionUniqueId uint64, errorCode int16) {
	//var response ErrorNtfPacket
	//sendBuf, _ := response.EncodingPacket(errorCode)
}

func ClientHeaderSize() int16 {
	return _ClientSessionHeaderSize
}

// Header의 PacketID만 읽는다.
func PeekPacketID(rawData []byte) int16 {
	packetID := binary.LittleEndian.Uint16(rawData[2:])
	return int16(packetID)
}

// 보디 데이터의 참조만 가져간다.
func PeekPacketBody(rawData []byte) (bodySize int16, refBody []byte) {
	headerSize := ClientHeaderSize()
	totalSize := int16(binary.LittleEndian.Uint16(rawData))
	bodySize = totalSize - headerSize

	if bodySize > 0 {
		refBody = rawData[headerSize:]
	}

	return bodySize, refBody
}


type LoginUserInfoNtfPacket struct {
	RoomNum int32
	UserId []byte
}

func (loginUserInfoNtf LoginUserInfoNtfPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _ClientSessionHeaderSize + 4 + 16
	sendBuf := make([]byte, totalSize)

	writer := MakeWriter(sendBuf, true)
	EncodingPacketHeader(&writer, totalSize, PACKET_ID_LOGIN_USER_INFO_NTF, 0)
	writer.WriteS32(loginUserInfoNtf.RoomNum)
	writer.WriteBytes(loginUserInfoNtf.UserId)
	return sendBuf, totalSize
}

// 나에게 다른 유저들의 정보를 전달해준다.
type LoginOtherUserInfoNtfPacket struct {
	TotalUserCount int32
	UserInfo        []LoginUserInfoNtfPacket
}

func (otherUserInfo LoginOtherUserInfoNtfPacket) EncodingPacket() ([]byte, int16) {
	totalSize := _ClientSessionHeaderSize + 4 + (4+16)*int16(otherUserInfo.TotalUserCount)
	sendBuf := make([]byte, totalSize)

	writer := MakeWriter(sendBuf , true)
	EncodingPacketHeader(&writer, totalSize , PACKET_ID_LOGIN_OTHER_USER_INFO_NTF , 0)
	writer.WriteS32(otherUserInfo.TotalUserCount)
	for i := 0 ; int32(i) < otherUserInfo.TotalUserCount ; i++ {
		writer.WriteS32(otherUserInfo.UserInfo[i].RoomNum)
		writer.WriteBytes(otherUserInfo.UserInfo[i].UserId)
	}
	return sendBuf , totalSize
}



