package gohipernetFake

import (
	"log"
	"net"
	"sync/atomic"
)

func init_Network_Impl(clientHaderSize int16 , serverHeaderSize int16) {
	defer PrintPanicStack()

	_InitNetworkSendFunction()
}

func start_Network_Impl(clientConfig *NetworkConfig, networkFunctor SessionNetworkFunctors) {
	defer PrintPanicStack()

	// 아래 함수가 호출되면 무한 대기에 들어간다.
	_tcpSessionManager = newClientSessionManager(clientConfig, networkFunctor)
	_start_TCPServer_block(clientConfig,networkFunctor)

}

func _start_TCPServer_block(config *NetworkConfig, networkFunctor SessionNetworkFunctors) {
	defer PrintPanicStack()
	Logger.Info("tcpServerStart - Start")
	IExportLog("Info" , "tcpServerStart - Start")

	var err error
	tcpAddr, _ := net.ResolveTCPAddr("tcp", config.BindAddress)
	_mClientListener, err = net.ListenTCP("tcp" , tcpAddr)

	if err != nil {
		log.Fatal("Error starting TCP server")
	}
	defer _mClientListener.Close()

	log.Println("Server Listen ...")

	for {
		conn , _ := _mClientListener.Accept()

		client := &TcpSession {
			SeqIndex: SeqNumIncrement(),
			TcpConn: conn ,
			NetworkFunctor:networkFunctor,
		}
		client.Index = int32(client.SeqIndex) - 1

		_tcpSessionManager.addSession(client)

		go client.handleTcpRead(networkFunctor)
	}

	Logger.Info("tcpServerStart - End")
	IExportLog("Info","tcpServerStart - End")
}

func _InitNetworkSendFunction() {
	NetLibSendToClinet = sendToClient
	NetLibSendToAllClient = sendToAllClinet
	NetLibPostSendToClient = postSendToClient
	NetLibPostSendToAllClient = postSendToAllClient

	Logger.Info("call _InitNetworkSendFunction")
}


func sendToClient(sessionIdex int32 , sessionUniqueID uint64, data []byte) bool {
	result := _tcpSessionManager.sendPacket(sessionIdex, sessionUniqueID , data)
	return result
}

func sendToAllClinet(sendData []byte) {
	_tcpSessionManager.sendPacketAllClient(sendData)
}


func postSendToClient(sessionIndex int32, sessionUniqueID uint64, data []byte) bool {
	return sendToClient(sessionIndex, sessionUniqueID , data)
}

func postSendToAllClient(sendData []byte) {
	_tcpSessionManager.sendPacketAllClient(sendData)
}



var _seqNumber uint64 // 절대 이것을 바로 사용하면 안 된다!!!

func SeqNumIncrement() uint64 {
	newValue := atomic.AddUint64(&_seqNumber , 1)
	return newValue
}

var _tcpSessionManager *tcpClientSessionManager
var _mClientListener *net.TCPListener
