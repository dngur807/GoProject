package main

import (
	. "study/basic_server/gohipernetFake"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type EchoServer struct {
	ServerIndex		int
	IP 				string
	Port			int
}

func createServer(netConfig NetworkConfig) {
	NTELIB_LOG_INFO("createServer!!!")

	var server EchoServer

	if server.setIPAddress(netConfig.BindAddress) == false {
		NTELIB_LOG_ERROR("fail server address")
		return
	}

	// 네트워크 함수 핸들링
	networkFunctor := SessionNetworkFunctors{}
	networkFunctor.OnConnect = server.OnConnect
	networkFunctor.OnReceive = server.OnReceive
	networkFunctor.OnReceiveBufferData = nil
	networkFunctor.OnClose = server.OnClose
	networkFunctor.PacketTotalSizeFunc = nil
	networkFunctor.PacketHeaderSize = PACKET_HEADER_SIZE
	networkFunctor.IsClientSession = true

	// 네트워크 관련 함수 핸들러
	NetLibInitNetwork(PACKET_HEADER_SIZE , PACKET_HEADER_SIZE)

	NetLibStartNetwork(&netConfig, networkFunctor)

}

func (server *EchoServer) setIPAddress (ipAddress string) bool {
	results := strings.Split(ipAddress, ":")
	if len(results) != 2 {
		return false
	}

	server.IP = results[0]
	server.Port, _ = strconv.Atoi(results[1])

	NTELIB_LOG_INFO("Server Address" , zap.String("IP", server.IP) , zap.Int("Port" , server.Port))
	return true
}

func (server *EchoServer) OnConnect(sessionIndex int32, sessionUniqueID uint64) {
	NTELIB_LOG_INFO("client OnConnect" , zap.Int32("sessionIndex",sessionIndex) , zap.Uint64("sessionUniqueId" , sessionUniqueID))
}

func (server *EchoServer) OnReceive(sessionIndex int32, sessionUniqueID uint64, data []byte) bool {
	NTELIB_LOG_DEBUG("OnReceive", zap.Int32("sessionIndex", sessionIndex),
		zap.Uint64("sessionUniqueID", sessionUniqueID),
		zap.Int("packetSize", len(data)))

	NetLibSendToClinet(sessionIndex, sessionUniqueID , data)
	return true
}

func (server *EchoServer) OnClose(sessionIndex int32, sessionUniqueID uint64) {
	NTELIB_LOG_INFO("client OnCloseClientSession" , zap.Int32("sessionIndex" , sessionIndex),  zap.Uint64("sessionUniqueId" , sessionUniqueID))
}