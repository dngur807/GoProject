package gohipernetFake

import (
	"encoding/binary"
	"errors"
	"reflect"
)

func packetTotalSize(data []byte) int16 {
	totalsize := binary.LittleEndian.Uint16(data)
	return int16(totalsize)
}

// 타입의 크기를 계산한다
func Sizeof(t reflect.Type) int {
	switch t.Kind() {
	case reflect.Array:
		//fmt.Println("reflect.Array")
		if s := Sizeof(t.Elem()); s >= 0 {
			return s * t.Len()
		}

	case reflect.Struct:
		//fmt.Println("reflect.Struct")
		sum := 0
		for i, n := 0, t.NumField(); i < n; i++ {
			s := Sizeof(t.Field(i).Type)
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		//fmt.Println("reflect.int")
		return int(t.Size())
	case reflect.Slice:
		//fmt.Println("reflect.Slice:", sizeof(t.Elem()))
		return 0
	}

	return -1
}

type RawPacketData struct {
	pos   int
	data  []byte
	order binary.ByteOrder
}

func MakeReader(buffer []byte, isLittleEndian bool) RawPacketData {
	if isLittleEndian {
		return RawPacketData{data: buffer, order: binary.LittleEndian}
	}
	return RawPacketData{data: buffer, order: binary.BigEndian}
}

func (p *RawPacketData) ReadS32() (ret int32, err error) {
	_ret, _err := p.ReadU32()
	ret = int32(_ret)
	err = _err
	return
}

func (p *RawPacketData) ReadU32() (ret uint32, err error) {
	if p.pos+4 > len(p.data) {
		err = errors.New("read uint32 failed")
		return
	}

	buf := p.data[p.pos : p.pos+4]
	ret = p.order.Uint32(buf)
	p.pos += 4
	return
}

func MakeWriter(buffer []byte, isLittleEndian bool) RawPacketData {
	if isLittleEndian {
		return RawPacketData{data: buffer, order: binary.LittleEndian}
	}
	return RawPacketData{data: buffer, order: binary.BigEndian}
}

func (p *RawPacketData) ReadBytes(readSize int) (refSlice []byte) {
	refSlice = p.data[p.pos : p.pos + readSize]
	p.pos += readSize
	return 
}


//=================================Writers
func (p *RawPacketData) WriteS8(v int8) {
	p.data[p.pos] = (byte)(v)
	p.pos++
}

func (p *RawPacketData) WriteU16(v uint16) {
	p.order.PutUint16(p.data[p.pos:], v)
	p.pos += 2
}

func (p *RawPacketData) WriteS16(v int16) {
	p.WriteU16(uint16(v))
}

func (p *RawPacketData) WriteBytes(v []byte) {
	copy(p.data[p.pos:], v)
	p.pos += len(v)
}

func (p *RawPacketData) WriteU32(v uint32) {
	p.order.PutUint32(p.data[p.pos:], v)
	p.pos += 4
}

func (p *RawPacketData) WriteS32(v int32) {
	p.WriteU32(uint32(v))
}

func (p *RawPacketData) WriteU64(v uint64) {
	p.order.PutUint64(p.data[p.pos:], v)
	p.pos += 8
}

func (p *RawPacketData) WriteS64(v int64) {
	p.WriteU64(uint64(v))
}

func (p *RawPacketData) WriteString(v string) {
	copyLen := copy(p.data[p.pos:], v)
	p.pos += copyLen
}


