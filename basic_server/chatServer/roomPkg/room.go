package roomPkg

import (
	. "GoStudy/basic_server/gohipernetFake"
	"GoStudy/basic_server/chatServer/protocol"
	"sync"
	"sync/atomic"
)

/**
sync.Pool은 일종의 메모리 풀이라고 볼 수 있다
자원을 풀에 넣었다가, 필요할때 다시 꺼내 쓰는 것이다.
 */
type baseRoom struct {
	_index               int32
	_number              int32 // 채널 고유 번호
	_config              RoomConfig
	_curUserCount        int32
	_roomUserUniqueIdSeq uint64
	_userPool            *sync.Pool

	// 자료구조를 배열로 바꾸는 것이 좋음
	_userSessionUniqueIdMap map[uint64]*roomUser // range 순회 시 복사 비용 발생해서 포인터 값을 사용한다.

	_funcPackIdlist []int16
	_funclist       []func(*roomUser, protocol.Packet) int16

	enterUserNotify func(int64, int32)
	leaveUserNotify func(int64)
}

func (room *baseRoom) initialize(index int32, config RoomConfig) {
	room._initialize(index, config)
	room._initUserPool()
	room._userSessionUniqueIdMap = make(map[uint64]*roomUser)
}

func (room *baseRoom) _initialize(index int32, config RoomConfig) {
	room._number = config.StartRoomNumber + index
	room._index = index
	room._config = config
}

func (room *baseRoom) _initUserPool() {
	room._userPool = &sync.Pool{
		New: func() interface{} {
			user := new(roomUser)
			return user
		},
	}
}
func (room *baseRoom) settingPacketFunction() {
	maxFuncListCount := 16
	room._funclist = make([]func(*roomUser, protocol.Packet) int16, 0, maxFuncListCount)
	room._funcPackIdlist = make([]int16, 0, maxFuncListCount)

	room._addPacketFunction(protocol.PACKET_ID_ROOM_ENTER_REQ, room._packetProcess_EnterUser)
	//room._addPacketFunction(protocol.PACKET_ID_ROOM_LEAVE_REQ, room._packetProcess_LeaveUser)
}

func (room *baseRoom) _addPacketFunction(packetID int16, packetFunc func(*roomUser, protocol.Packet) int16) {
	room._funclist = append(room._funclist, packetFunc)
	room._funcPackIdlist = append(room._funcPackIdlist, packetID)
}

func (room *baseRoom) addUser(userInfo addRoomUserInfo) (*roomUser, int16) {
	if room._IsFullUser() {
		return nil, protocol.ERROR_CODE_ENTER_ROOM_USER_FULL
	}
	if room.getUser(userInfo.netSessionUniqueId) != nil {
		return nil, protocol.ERROR_CODE_ENTER_ROOM_DUPLCATION_USER
	}
	atomic.AddInt32(&room._curUserCount, 1)

	user := room._getUserObject()
	user.init(userInfo.userID, room.generateUserUniqueId())
	user.SetNetworkInfo(userInfo.netSessionIndex, userInfo.netSessionUniqueId)
	user.packetDataSize = user.PacketDataSize()

	room._userSessionUniqueIdMap[user.netSessionUniqueId] = user
	return user, protocol.ERROR_CODE_NONE
}

func (room *baseRoom) _IsFullUser() bool {
	if room.getCurUserCount() == room._config.MaxUserCount {
		return true
	}
	return false
}

func (room *baseRoom) getCurUserCount() int32 {
	count := atomic.LoadInt32(&room._curUserCount)
	return count
}

func (room *baseRoom) getNumber() int32 {
	return room._number
}

func (room *baseRoom) getUser(sessionUniqueId uint64) *roomUser {
	if user, ok := room._userSessionUniqueIdMap[sessionUniqueId]; ok {
		return user
	}
	return nil
}

func (room *baseRoom) _getUserObject() *roomUser {
	userObject := room._userPool.Get().(*roomUser)
	return userObject
}

func (room *baseRoom) generateUserUniqueId() uint64 {
	room._roomUserUniqueIdSeq++
	uniqueId := room._roomUserUniqueIdSeq
	return uniqueId
}

// 유저 하나에게 보낼 때는 통으로 보낸다.
func (room *baseRoom) _allocUserInfo(user *roomUser) (dataSize int16, dataBuffer []byte) {
	dataSize = user.packetDataSize
	dataBuffer = make([]byte, dataSize)
	writer := MakeWriter(dataBuffer, true)
	_writeUserInfo(&writer, user)

	return dataSize, dataBuffer
}
func _writeUserInfo(writer *RawPacketData, user *roomUser) {
	writer.WriteU64(user.RoomUniqueId)
	writer.WriteS8(user.IDLen)
	writer.WriteBytes(user.ID[0:user.IDLen])
}

func (room *baseRoom) broadcastPacket(packetSize int16,
	sendPacket []byte,
	exceptSessionUniqueId uint64) {

	for _, user := range room._userSessionUniqueIdMap {
		if user.netSessionUniqueId == exceptSessionUniqueId {
			continue
		}
		NetLibPostSendToClient(user.netSessionIndex, user.netSessionUniqueId, sendPacket)
	}
}

func (room *baseRoom) allocAllUserInfo(exceptSessionUniqueId uint64) (userCount int8, dataSize int16, dataBuffer []byte) {
	for _, user := range room._userSessionUniqueIdMap {
		if user.netSessionUniqueId == exceptSessionUniqueId {
			continue
		}

		userCount++
		dataSize += user.packetDataSize
	}

	dataBuffer = make([]byte, dataSize)
	writer := MakeWriter(dataBuffer, true)

	for _, user := range room._userSessionUniqueIdMap {
		if user.netSessionUniqueId == exceptSessionUniqueId {
			continue
		}

		_writeUserInfo(&writer, user)
	}
	return userCount, dataSize, dataBuffer
}
