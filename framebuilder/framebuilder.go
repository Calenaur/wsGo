package framebuilder

import (
	"bytes"

	"github.com/Calenaur/wsGo/frame"
)

const STAGE_BASE = 0
const STAGE_LENGTH16 = 1
const STAGE_LENGTH64 = 2
const STAGE_MASK = 3
const STAGE_DATA = 4

const BYTE_OPCODE = 0
const BYTE_LENGTH = 1

type FrameBuilder struct {
	Index           uint64
	StageStartIndex uint64
	Available       int
	Stage           int
	buffer          *bytes.Buffer
	frames          []*frame.Frame
	current         *frame.Frame
	finished        bool
}

func New() *FrameBuilder {
	return &FrameBuilder{0, 0, 0, 0, bytes.NewBuffer([]byte{}), []*frame.Frame{}, frame.New(), false}
}

func (fb *FrameBuilder) Write(b byte) {
	var f *frame.Frame = fb.current
	switch fb.Stage {
	case STAGE_BASE:
		switch fb.Index {
		case BYTE_OPCODE:
			f.ParseFlags(b).ParseOpcode(b)
		case BYTE_LENGTH:
			fb.handleLength(f, b)
		}
	case STAGE_LENGTH16:
		fb.handleLength16(f, b)
	case STAGE_LENGTH64:
		fb.handleLength64(f, b)
	case STAGE_MASK:
		fb.handleMask(f, b)
	case STAGE_DATA:
		fb.handleData(f, b)
	}
	if fb.finished {
		fb.finished = false
		return
	}
	fb.Index++
}

func (fb *FrameBuilder) handleLength(f *frame.Frame, b byte) {
	f.ParseMask(b).ParseLength(b)
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
	fb.buffer.Reset()
}

func (fb *FrameBuilder) handleLength16(f *frame.Frame, b byte) {
	fb.buffer.WriteByte(b)
	if fb.Index >= fb.StageStartIndex+2 {
		f.ParseLength16(fb.buffer.Bytes())
		if f.Masked {
			fb.Stage = STAGE_MASK
		} else {
			fb.Stage = STAGE_DATA
		}
		fb.StageStartIndex = fb.Index
		fb.buffer.Reset()
	}
}

func (fb *FrameBuilder) handleLength64(f *frame.Frame, b byte) {
	fb.buffer.WriteByte(b)
	if fb.Index >= fb.StageStartIndex+8 {
		f.ParseLength64(fb.buffer.Bytes())
		if f.Masked {
			fb.Stage = STAGE_MASK
		} else {
			fb.Stage = STAGE_DATA
		}
		fb.StageStartIndex = fb.Index
		fb.buffer.Reset()
	}
}

func (fb *FrameBuilder) handleMask(f *frame.Frame, b byte) {
	fb.buffer.WriteByte(b)
	if fb.Index >= fb.StageStartIndex+4 {
		if f.Length == 0 {
			f.Data = []byte{}
			fb.finish(f)
			return
		}
		f.Mask = make([]byte, fb.buffer.Len())
		copy(f.Mask, fb.buffer.Bytes())
		fb.Stage = STAGE_DATA
		fb.StageStartIndex = fb.Index
		fb.buffer.Reset()
	}
}

func (fb *FrameBuilder) handleData(f *frame.Frame, b byte) {
	if f.Length == 0 {
		f.Data = []byte{}
		fb.finish(f)
		return
	} else {
		fb.buffer.WriteByte(b)
		if fb.Index >= fb.StageStartIndex+f.Length {
			f.Data = make([]byte, fb.buffer.Len())
			copy(f.Data, fb.buffer.Bytes())
			fb.finish(f)
			return
		}
	}
}

func (fb *FrameBuilder) finish(f *frame.Frame) {
	if f.Masked {
		f.Unmask()
	}
	fb.frames = append(fb.frames, f)
	fb.Available++
	fb.Clear()
	fb.finished = true
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

func (fb *FrameBuilder) Clear() {
	fb.Index = 0
	fb.StageStartIndex = 0
	fb.Stage = STAGE_BASE
	fb.buffer = bytes.NewBuffer([]byte{})
	fb.current = frame.New()
}

func (fb *FrameBuilder) Reset() {
	fb.Index = 0
	fb.StageStartIndex = 0
	fb.Available = 0
	fb.Stage = STAGE_BASE
	fb.buffer = bytes.NewBuffer([]byte{})
	fb.frames = []*frame.Frame{}
	fb.current = frame.New()
	fb.finished = false
}
