package gohipernetFake

func NetLibInitLog() {
	init_Log()
	wrapLoggerFunc()
}

// 네트워크 초기화
func NetLibInitNetwork(clientHeaderSize int16 , serverHeaderSize int16) {
	init_Network_Impl(clientHeaderSize , serverHeaderSize)
}

// 네트워크 시작
func NetLibStartNetwork(clientConfig *NetworkConfig, networkFunctor SessionNetworkFunctors) {
	start_Network_Impl(clientConfig , networkFunctor)
}

// 특정 클라이언트에게 데이터를 보낸다
var NetLibSendToClinet func(int32 , uint64, []byte)bool
// 접속된 모든 클라이언트에게 데이터를 보낸다.
var NetLibSendToAllClient func([]byte)

var NetLibPostSendToAllClient func([]byte)
var NetLibPostSendToClient func(int32, uint64, []byte)bool
