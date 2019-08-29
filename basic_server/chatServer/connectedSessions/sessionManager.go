package connectedSessions

import (
	"GoStudy/basic_server/chatServer/protocol"
	. "GoStudy/basic_server/gohipernetFake"
	"bytes"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
)

var _manager Manager

// 스레드 세이프 해야 한다.
type Manager struct {
	_UserIDsessionMap      *sync.Map
	_maxSessionCount       int32
	_sessionList           []*session
	_maxUserCount          int32
	_currentLoginUserCount int32
}

func Init(maxSessionCount int, maxUserCount int32) bool {
	_manager._UserIDsessionMap = new(sync.Map)
	_manager._maxUserCount = maxUserCount
	_manager._maxSessionCount = int32(maxSessionCount)
	_manager._sessionList = make([]*session, maxSessionCount)
	_manager._currentLoginUserCount = 0

	for i := 0; i < maxSessionCount; i++ {
		_manager._sessionList[i] = new(session)
		index := int32(i)
		_manager._sessionList[i].Init(index)
	}
	return true
}

func AddSession(sessionIndex int32, sessionUniqueID uint64) bool {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex", sessionIndex))
		return false
	}

	if _manager._sessionList[sessionIndex].GetConnectTimeSec() > 0 {
		NTELIB_LOG_ERROR("already connected session", zap.Int32("sessionIndex", sessionIndex))
		return false
	}

	// 방어적인 목적으로 한번 더 clear를 한다.
	_manager._sessionList[sessionIndex].Clear()
	//_manager._sessionList[sessionIndex].SetConnectTimeSec(NetLib_GetCurrentUnixTime(), sessionUniqueID)
	return true
}

func _validSessionIndex(index int32) bool {
	if index < 0 || index >= _manager._maxSessionCount {
		return false
	}
	return true
}

func IsLoginUser(sessionIndex int32) bool {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex", sessionIndex))
		return false
	}
	return _manager._sessionList[sessionIndex].IsAuth()
}

func RemoveSession(sessionIndex int32, isLoginUser bool) bool {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex", sessionIndex))
		return false
	}

	if isLoginUser {
		atomic.AddInt32(&_manager._currentLoginUserCount, -1)

		userID := string(_manager._sessionList[sessionIndex].getUserID())
		_manager._UserIDsessionMap.Delete(userID)
	}

	_manager._sessionList[sessionIndex].Clear()
	
	return true
}
func GetCurrentUserCount() int32 {
	return _manager._currentLoginUserCount
}

func GetUserIDList(exUserId []byte) (protocol.LoginOtherUserInfoNtfPacket, bool)  {
	var UserInfoList  protocol.LoginOtherUserInfoNtfPacket
	var index int16

	UserInfoList.TotalUserCount = _manager._currentLoginUserCount - 1
	UserInfoList.UserInfo = make([]protocol.LoginUserInfoNtfPacket , UserInfoList.TotalUserCount)
	_manager._UserIDsessionMap.Range(func(k, v interface{}) bool {
		session := v.(*session)

		userId := bytes.Trim(session._userID[:], "\x00")

		// 예외 아이디 제외
		if string(userId) ==  string(exUserId) {
			return true
		}

		UserInfoList.UserInfo[index].UserId = session._userID[:]
		UserInfoList.UserInfo[index].RoomNum = session._RoomNum
		index++
		return true
	})
	return UserInfoList, true
}

func GetUserID(sessionIndex int32) ([]byte, bool) {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex", sessionIndex))
		return nil, false
	}

	return _manager._sessionList[sessionIndex].getUserID(), true
}

func SetRoomNumber(sessionIndex int32, sessionUniqueId uint64, roomNum int32, curTimeSec int64) bool {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex", sessionIndex))
		return false
	}
	return _manager._sessionList[sessionIndex].setRoomNumber(sessionUniqueId, roomNum, curTimeSec)
}

func GetRoomNumber(sessionIndex int32) (int32, int32) {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex", sessionIndex))
		return -1, -1
	}
	return _manager._sessionList[sessionIndex].getRoomNumber()
}

func SetLogin(sessionIndex int32, sessionUniqueId uint64, userID []byte, curTimeSec int64) bool {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex", sessionIndex))
		return false
	}

	newUserID := string(userID)

	if _, ok := _manager._UserIDsessionMap.Load(newUserID); ok {
		return false
	}

	_manager._sessionList[sessionIndex].SetUser(sessionUniqueId, userID, curTimeSec)
	_manager._UserIDsessionMap.Store(newUserID, _manager._sessionList[sessionIndex])

	atomic.AddInt32(&_manager._currentLoginUserCount, 1)
	return true
}
