package connectedSessions

import (
	"go.uber.org/zap"
	. "study/basic_server/gohipernetFake"
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
	
	for i := 0 ; i < maxSessionCount ; i++ {
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
	
	if _manager._sessionList[sessionIndex].GetConnectTimeSec() > 0  {
		NTELIB_LOG_ERROR("already connected session", zap.Int32("sessionIndex", sessionIndex))
		return false
	}
	
	// 방어적인 목적으로 한번 더 clear를 한다.
	_manager._sessionList[sessionIndex].Clear()
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
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex" , sessionIndex))
		return false
	}
	return _manager._sessionList[sessionIndex].IsAuth()
}

func RemoveSession(sessionIndex int32, isLoginUser bool) bool {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex" , zap.Int32("sessionIndex" , sessionIndex))
		return false
	}
	
	if isLoginUser {
		atomic.AddInt32(&_manager._currentLoginUserCount , -1)
		
		userID := string(_manager._sessionList[sessionIndex].getUserID())
		_manager._UserIDsessionMap.Delete(userID)
	}
	
	_manager._sessionList[sessionIndex].Clear()
	return true
}

func GetUserID(sessionIndex int32) ([]byte, bool) {
	if _validSessionIndex(sessionIndex) == false {
		NTELIB_LOG_ERROR("Invalid sessionIndex", zap.Int32("sessionIndex" , sessionIndex))
		return nil, false
	}
	
	return _manager._sessionList[sessionIndex].getUserID(), true
}