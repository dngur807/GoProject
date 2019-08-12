package roomPkg

import (
	"go.uber.org/zap"
	. "study/basic_server/gohipernetFake"
)

type RoomManager struct {
	_roomStartNum  int32
	_maxRoomCount  int32
	_roomCountList []int16
	_roomList      []baseRoom
}

func NewRoomManager(config RoomConfig) *RoomManager {
	roomManager := new(RoomManager)
	roomManager._initialize(config)
	return roomManager
}

func (roomMgr *RoomManager) _initialize(config RoomConfig) {
	roomMgr._roomStartNum = config.StartRoomNumber
	roomMgr._maxRoomCount = config.MaxRoomCount
	roomMgr._roomCountList = make([]int16, config.MaxRoomCount)
	roomMgr._roomList = make([]baseRoom, config.MaxRoomCount)

	for i := int32(0); i < roomMgr._maxRoomCount; i++ {
		roomMgr._roomList[i].initialize(i, config)
		roomMgr._roomList[i].settingPacketFunction()
	}

	_log_StartRoomPacketProcess(config.MaxRoomCount, config)
	NTELIB_LOG_INFO("[[[RoomManager initialize - Park]]]", zap.Int32("_maxRoomCount", roomMgr._maxRoomCount))
}

func _log_StartRoomPacketProcess(maxRoomCount int32, config RoomConfig) {
	NTELIB_LOG_INFO("[[[RoomManager _startRoomPacketProcess]]]",
		zap.Int32("maxRoomCount", maxRoomCount),
		zap.Int32("StartRoomNumber", config.StartRoomNumber),
		zap.Int32("MaxUserCount", config.MaxUserCount))
}