package main

import (
	. "study/basic_server/gohipernetFake"
)

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

	//for {
	//	packet := <-server.PacketChan
	//	sessionIndex := packet.UserSessionIndex
	//	sessionUniqueId := packet.UserSessionUniqueId
	//	bodySize := packet.DataSize
	//	bodyData := packet.Data
	//
	//	//if packet.Id == protocol.PACKET_ID_LOGIN_REQ {
	//	//	ProcessPacketLogin(sessionIndex, sessionUniqueId, bodySize, bodyData)
	//	//} else if packet.Id == protocol.PACKET_ID_SESSION_CLOSE_SYS {
	//	//	ProcessPacketSessionClosed(server, sessionIndex, sessionUniqueId)
	//	//} else {
	//	//	roomNumber , _ := connectSessions.GetRoomNUmber(sessionIndex)
	//	//}
	//}
	//
	return IsWantedTermination
}
