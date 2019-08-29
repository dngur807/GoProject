package main

import (
	"GoStudy/basic_server/chatServer/connectedSessions"
	"GoStudy/basic_server/chatServer/protocol"
	. "GoStudy/basic_server/gohipernetFake"
	"bytes"
	"go.uber.org/zap"
)

func (server *ChatServer) DistributePacket(sessionIndex int32,
	sessionUniqueId uint64,
	packetData []byte) {

	packetID := protocol.PeekPacketID(packetData)
	bodySize, bodyData := protocol.PeekPacketBody(packetData)
	NTELIB_LOG_DEBUG("DistributePacket", zap.Int32("sessionIndex", sessionIndex), zap.Uint64("sessionUniqueId", sessionUniqueId), zap.Int16("PacketID", packetID))

	packet := protocol.Packet{Id: packetID}
	packet.UserSessionIndex = sessionIndex
	packet.UserSessionUniqueId = sessionUniqueId
	packet.Id = packetID
	packet.DataSize = bodySize
	packet.Data = make([]byte, packet.DataSize)
	copy(packet.Data, bodyData)

	server.PacketChan <- packet
	NTELIB_LOG_DEBUG("_distributePacket", zap.Int32("sessionIndex", sessionIndex), zap.Int16("PacketId", packetID))
}

func (server *ChatServer) PacketProcess_goroutine() {
	NTELIB_LOG_INFO("start PacketProcess goroutine")

	for {
		if server.PacketProcess_goroutine_Impl() {
			NTELIB_LOG_INFO("Wanted Stop PacketProcess goroutine")
			break
		}
	}
	NTELIB_LOG_INFO("Stop rooms PacketProcess goroutine")
}

func (server *ChatServer) PacketProcess_goroutine_Impl() bool {
	IsWantedTermination := false
	defer PrintPanicStack()

	for {
		packet := <-server.PacketChan
		sessionIndex := packet.UserSessionIndex
		sessionUniqueId := packet.UserSessionUniqueId
		bodySize := packet.DataSize
		bodyData := packet.Data

		if packet.Id == protocol.PACKET_ID_LOGIN_REQ {
			ProcessPacketLogin(sessionIndex, sessionUniqueId, bodySize, bodyData)
		} else if packet.Id == protocol.PACKET_ID_SESSION_CLOSE_SYS {
			ProcessPacketSessionClosed(server, sessionIndex, sessionUniqueId)
		} else {
			var requestPacket protocol.RoomEnterReqPacket
			(&requestPacket).Decoding(packet.Data)

			roomNumber, _ := connectedSessions.GetRoomNumber(sessionIndex)
			if (roomNumber == -1 ) {
				roomNumber = requestPacket.RoomNumber
			}
			server.RoomMgr.PacketProcess(roomNumber, packet)
		}
	}
	return IsWantedTermination
}

func ProcessPacketLogin(sessionIndex int32,
	sessionUniqueId uint64,
	bodySize int16,
	bodyData []byte) {

	// DB와 연동하지 않으므로 중복 로그인만 아니면 다 성공으로 한다
	var request protocol.LoginReqPacket
	if (&request).Decoding(bodyData) == false {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_PACKET_DECODING_FAIL)
		return
	}

	userId := bytes.Trim(request.UserID[:], "\x00")

	if len(userId) <= 0 {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_LOGIN_USER_INVALID_ID)
		return
	}

	curTime := NetLib_GetCurrentUnixTime()

	if connectedSessions.SetLogin(sessionIndex, sessionUniqueId, userId, curTime) == false {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_LOGIN_USER_DUPLICATION)
		return
	}

	_sendLoginResult(sessionIndex, sessionUniqueId , protocol.ERROR_CODE_NONE)

	// 접속한 유저에게 나에 접속을 알려준다.
	_sendAllLoginUserInfoNotify(sessionIndex, sessionUniqueId)
	// 나에게 다른 유저 정보를 알려준다.
	_sendMeLoginOtherUserInfoNotify(sessionIndex, sessionUniqueId , userId)
}

func _sendLoginResult(sessionIndex int32, sessionUniqueId uint64, result int16) {
	var response protocol.LoginResPacket
	response.Result = result
	sendPacket, _ := response.EncodingPacket()
	NetLibPostSendToClient(sessionIndex, sessionUniqueId, sendPacket)
	NTELIB_LOG_DEBUG("SendLoginResult", zap.Int32("sessionIndex", sessionIndex), zap.Int16("result", result))
}

func _sendAllLoginUserInfoNotify(sessionIndex int32, sessionUniqueId uint64) {
	login_user_id , _ := connectedSessions.GetUserID(sessionIndex)

	// 유저 아이디 존재 X
	if login_user_id == nil {
		NTELIB_LOG_ERROR("sendLoginUserInfoNotify - UserID not exist", zap.Int32("sessionIndex", sessionIndex))
	}
	
	// 접속한 모든 유저에게 접속 정보 전달해준다.
	var response protocol.LoginUserInfoNtfPacket
	response.RoomNum , _ = connectedSessions.GetRoomNumber(sessionIndex)
	response.UserId = login_user_id
	sendPacket, _ := response.EncodingPacket()
	NetLibPostSendToAllClient(sendPacket)
	NTELIB_LOG_DEBUG("Notify LoginUserInfo", zap.Int32("sessionIndex", sessionIndex), zap.String("UserId", string(response.UserId)))
}

func _sendMeLoginOtherUserInfoNotify(sessionIndex int32, sessionUniqueId uint64 , userId []byte) {
	var response protocol.LoginOtherUserInfoNtfPacket
	var result bool

	response , result = connectedSessions.GetUserIDList(userId)
	if result == false {
		NTELIB_LOG_DEBUG("Notify Empty OtherUser", zap.Int32("CurUserCount", connectedSessions.GetCurrentUserCount()))
		return
	}

	// 나에게 다른 유저들의 정보를 전달해준다.
	sendPacket , _ := response.EncodingPacket()
	NetLibPostSendToClient(sessionIndex, sessionUniqueId , sendPacket)
	NTELIB_LOG_DEBUG("Notify LoginOtherUserInfo", zap.Int32("sessionIndex", sessionIndex))
}


func ProcessPacketSessionClosed(server *ChatServer, sessionIndex int32, sessionUniqueId uint64) {
	roomNumber, _ := connectedSessions.GetRoomNumber(sessionIndex)

	if roomNumber > -1 {
		packet := protocol.Packet{
			sessionIndex,
			sessionUniqueId,
			protocol.PACKET_ID_ROOM_LEAVE_REQ,
			0,
			nil,
		}

		server.RoomMgr.PacketProcess(roomNumber, packet)
	}

	connectedSessions.RemoveSession(sessionIndex, true)
}
