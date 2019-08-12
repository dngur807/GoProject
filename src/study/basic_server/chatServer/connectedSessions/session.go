package connectedSessions

import (
	"study/basic_server/chatServer/protocol"
	"sync/atomic"
)

type session struct {
	_index             int32
	_networkUniqueID   uint64 // 네트워크 세션의 유니크 ID
	_userID            [protocol.MAX_USER_ID_BYTE_LENGTH]byte
	_userIDLength      int8
	_connectTimeSec    int64 // 연결된 시간
	_RoomNum           int32
	_RoomNumOfEntering int32 // 현재 입장 중인 룸의 번호
}

func (session *session) Init(index int32) {
	session._index = index
	session.Clear()
}

func (session *session) Clear() {
	session._ClearUserId()
	session.setRoomNumber(0, -1, 0)
}
func (session *session) _ClearUserId() {
	session._userIDLength = 0
}

func (session *session) SetConnectTimeSec(timeSec int64, uniqueID uint64) {

}

func (session *session) GetConnectTimeSec() int64 {
	return atomic.LoadInt64(&session._connectTimeSec)
}

func (session *session) setRoomNumber(sessionUniqueId uint64, roomNum int32, curTimeSec int64) bool {
	if roomNum == -1 {
		atomic.StoreInt32(&session._RoomNum, roomNum)
		atomic.StoreInt32(&session._RoomNumOfEntering, roomNum)
		return true
	}

	if sessionUniqueId != 0 && session.validNetworkUniqueID(sessionUniqueId) == false {
		return false
	}

	// 입력이력 -1이 아닌경우
	// -1이 아닐 때만 cas으로 변경한다.  실패하면 채널 입장도 실패한다.
	if atomic.CompareAndSwapInt32(&session._RoomNum, -1, roomNum) == false {
		return false
	}

	atomic.StoreInt32(&session._RoomNumOfEntering, roomNum)
	return true
}

func (session *session) validNetworkUniqueID(sessionUniqueId uint64) bool {
	return atomic.LoadUint64(&session._networkUniqueID) == sessionUniqueId
}

func (session *session) IsAuth() bool {
	if session._userIDLength > 0 {
		return true
	}
	return false
}

func (session *session) getUserID() []byte {
	return session._userID[0:session._userIDLength]
}
