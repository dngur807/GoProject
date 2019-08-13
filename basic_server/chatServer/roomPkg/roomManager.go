package roomPkg

import (
	"GoStudy/basic_server/chatServer/protocol"
	. "GoStudy/basic_server/gohipernetFake"
	"go.uber.org/zap"
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

func (roomMgr *RoomManager) PacketProcess(roomNumber int32, packet protocol.Packet) {
	NTELIB_LOG_DEBUG("[[RoomManager - PacketProcess]]", zap.Int16("PacketID", packet.Id))
	
	room := roomMgr.getRoomByNumber(roomNumber)
	if room == nil {
		protocol.NotifyErrorPacket(packet.UserSessionIndex, packet.UserSessionUniqueId, protocol.ERROR_CODE_ROOM_INVALIDE_NUMBER)
		return
	}

	user := room.getUser(packet.UserSessionUniqueId)

	if user == nil && packet.Id != protocol.PACKET_ID_ROOM_ENTER_REQ {
		protocol.NotifyErrorPacket(packet.UserSessionIndex, packet.UserSessionUniqueId, protocol.ERROR_CODE_ROOM_NOT_IN_USER)
		return
	}

	funcCount := len(room._funcPackIdlist)
	for i := 0; i < funcCount; i++ {
		if room._funcPackIdlist[i] != packet.Id {
			continue
		}

		result := room._funclist[i](user, packet)
		if result != protocol.ERROR_CODE_NONE {
			NTELIB_LOG_DEBUG("[[Room - _packetProcess - Fail]]",
				zap.Int16("PacketID", packet.Id), zap.Int16("Error", result))
		}
		return
	}
	NTELIB_LOG_DEBUG("[[Room - _packetProcess - Fail(Not Registered)]]", zap.Int16("PacketID", packet.Id))
}

func (roomMgr *RoomManager) getRoomByNumber(roomNumber int32) *baseRoom {
	roomIndex := roomNumber - roomMgr._roomStartNum

	if roomNumber < 0 || roomIndex >= roomMgr._maxRoomCount {
		return nil
	}
	return &roomMgr._roomList[roomIndex]
}
