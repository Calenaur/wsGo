package frame

import (
	"bytes"
	"encoding/binary"
)

const OPCODE_CONT = 0
const OPCODE_TEXT = 1
const OPCODE_BINARY = 2
const OPCODE_CLOSE = 8
const OPCODE_PING = 9
const OPCODE_PONG = 10

type Frame struct {
	Fin    bool
	Rsv1   bool
	Rsv2   bool
	Rsv3   bool
	Opcode int
	Masked bool
	Length uint64
	Mask   []byte
	Data   []byte
}

func New() *Frame {
	return &Frame{false, false, false, false, 0, false, 0, []byte{}, []byte{}}
}

func NewFromString(s string) *Frame {
	f := New()
	f.Fin = true
	f.Opcode = 1
	f.Masked = false
	f.Data = []byte(s)
	f.Length = uint64(len(f.Data))
	return f
}

func NewFromBinary(opcode int, data []byte) *Frame {
	f := New()
	f.Fin = true
	f.Opcode = opcode
	f.Masked = false
	f.Data = data
	f.Length = uint64(len(f.Data))
	return f
}

func (f *Frame) ToBytes() []byte {
	var b byte
	buf := bytes.NewBuffer(nil)
	if f.Fin {
		b = b | 128
	}
	if f.Rsv1 {
		b = b | 64
	}
	if f.Rsv2 {
		b = b | 32
	}
	if f.Rsv3 {
		b = b | 16
	}
	buf.WriteByte(b | byte(f.Opcode))
	if f.Masked {
		b = 128
	} else {
		b = 0
	}
	if f.Length > 125 && f.Length <= 65535 {
		a := make([]byte, 8)
		binary.LittleEndian.PutUint16(a, uint16(f.Length))
		buf.WriteByte(b | 126)
		buf.Write(a)
	} else if f.Length > 65535 {
		a := make([]byte, 2)
		binary.LittleEndian.PutUint64(a, f.Length)
		buf.WriteByte(b | 127)
		buf.Write(a)
	} else {
		buf.WriteByte(b | byte(f.Length))
	}
	if f.Masked {
		buf.Write(f.Mask)
	}
	buf.Write(f.Data)
	return buf.Bytes()
}

func (f *Frame) Unmask() {
	if f.Masked {
		var i uint64 = 0
		for i < f.Length {
			f.Data[i] = f.Data[i] ^ f.Mask[i%4]
			i++
		}
		f.Masked = false
	}
}
