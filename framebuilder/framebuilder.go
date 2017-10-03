package framebuilder

import (
	"encoding/binary"

	"github.com/Calenaur/wsGo/frame"
)

const STAGE_BASE = 0
const STAGE_LENGTH16 = 1
const STAGE_LENGTH64 = 2
const STAGE_MASK = 3
const STAGE_DATA = 4

type FrameBuilder struct {
	Index           uint64
	StageStartIndex uint64
	Available       int
	Stage           int
	buffer          []byte
	frames          []*frame.Frame
	current         *frame.Frame
}

func New() *FrameBuilder {
	return &FrameBuilder{0, 0, 0, 0, []byte{}, []*frame.Frame{}, frame.New()}
}

func (fb *FrameBuilder) Write(b byte) {
	var f *frame.Frame = fb.current
	switch fb.Stage {
	case STAGE_BASE:
		switch fb.Index {
		case 0:
			f.Fin = b&128 == 128
			f.Rsv1 = b&64 == 64
			f.Rsv2 = b&32 == 32
			f.Rsv3 = b&16 == 16
			f.Opcode = int(b & 15)
		case 1:
			f.Masked = b&128 == 128
			f.Length = uint64(b & 127)
			if f.Length < 126 {
				if f.Masked {
					fb.Stage = STAGE_MASK
				} else {
					fb.Stage = STAGE_DATA
				}
			} else if f.Length == 126 {
				fb.Stage = STAGE_LENGTH16
			} else if f.Length == 127 {
				fb.Stage = STAGE_LENGTH64
			}
			fb.StageStartIndex = fb.Index
			fb.clearBuffer()
		}
	case STAGE_LENGTH16:
		fb.buffer = append(fb.buffer, b)
		if fb.Index >= fb.StageStartIndex+2 {
			f.Length = uint64(binary.LittleEndian.Uint16(fb.buffer))
			if f.Masked {
				fb.Stage = STAGE_MASK
			} else {
				fb.Stage = STAGE_DATA
			}
			fb.StageStartIndex = fb.Index
			fb.clearBuffer()
		}
	case STAGE_LENGTH64:
		fb.buffer = append(fb.buffer, b)
		if fb.Index >= fb.StageStartIndex+8 {
			f.Length = binary.LittleEndian.Uint64(fb.buffer)
			if f.Masked {
				fb.Stage = STAGE_MASK
			} else {
				fb.Stage = STAGE_DATA
			}
			fb.StageStartIndex = fb.Index
			fb.clearBuffer()
		}
	case STAGE_MASK:
		fb.buffer = append(fb.buffer, b)
		if fb.Index >= fb.StageStartIndex+4 {
			if f.Length == 0 {
				f.Data = []byte{}
				fb.finish(f)
				return
			}
			f.Mask = fb.buffer
			fb.Stage = STAGE_DATA
			fb.StageStartIndex = fb.Index
			fb.clearBuffer()
		}
	case STAGE_DATA:
		if f.Length == 0 {
			f.Data = []byte{}
			fb.finish(f)
			return
		} else {
			fb.buffer = append(fb.buffer, b)
			if fb.Index >= fb.StageStartIndex+f.Length {
				f.Data = fb.buffer
				fb.finish(f)
				return
			}
		}
	}
	fb.Index++
}

func (fb *FrameBuilder) finish(f *frame.Frame) {
	if f.Masked {
		f.Unmask()
	}
	fb.frames = append(fb.frames, f)
	fb.Available++
	fb.Clear()
}

func (fb *FrameBuilder) TakeFrame() *frame.Frame {
	var f *frame.Frame
	if len(fb.frames) > 0 {
		f = fb.frames[0]
		fb.frames = append(fb.frames[:0], fb.frames[1:]...)
		fb.Available--
	}
	return f
}

func (fb *FrameBuilder) clearBuffer() {
	fb.buffer = []byte{}
}

func (fb *FrameBuilder) Clear() {
	fb.Index = 0
	fb.StageStartIndex = 0
	fb.Stage = STAGE_BASE
	fb.buffer = []byte{}
	fb.current = frame.New()
}

func (fb *FrameBuilder) Reset() {
	fb.Index = 0
	fb.StageStartIndex = 0
	fb.Available = 0
	fb.Stage = STAGE_BASE
	fb.buffer = []byte{}
	fb.frames = []*frame.Frame{}
	fb.current = frame.New()
}
