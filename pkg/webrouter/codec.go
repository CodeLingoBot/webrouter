package webrouter

import (
	"bytes"
	"encoding/json"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/morya/net"
	"github.com/funny/link"
	"io"
)

type codecFactory struct {
	backend *Backend
}

func newCodecFactory(backend *Backend) link.CodecType {
	return codecFactory{
		backend: backend,
	}
}

func (self codecFactory) NewEncoder(w io.Writer) link.Encoder {
	return myEncoder{
		w:       w,
		backend: self.backend,
	}
}

func (self codecFactory) NewDecoder(r io.Reader) link.Decoder {
	return myDecoder{
		r:       r,
		buf:     net.NewByteBuf(128),
		backend: self.backend,
		length:  8,
	}
}

type myEncoder struct {
	w       io.Writer
	backend *Backend
}

type myDecoder struct {
	r       io.Reader
	buf     *net.ByteBuf
	length  int
	backend *Backend
}

func (self myEncoder) Encode(msg interface{}) error {
	var err error

	if value, ok := msg.([]byte); ok {
		err = self.writeBytes(value)
	} else if value, ok := msg.(byte); ok {
		err = self.writeByte(value)
	} else if value, ok := msg.(int); ok {
		err = self.writeInt(value)
	} else if value, ok := msg.(string); ok {
		self.writeInt(int(len(value)))
		self.writeByte(self.backend.msgType)
		self.writeByte(byte(0))
		self.writeByte(0x0d)
		self.writeByte(0x0a)
		err = self.writeBytes([]byte(value))
	}

	return err
}

func (self myEncoder) writeInt(value int) error {
	_, err := self.w.Write(intToBytes(value))
	return err
}

func (self myEncoder) writeByte(value byte) error {
	_, err := self.w.Write([]byte{value})
	return err
}

func (self myEncoder) writeBytes(value []byte) error {
	_, err := self.w.Write(value)
	return err
}

func (self myDecoder) Decode(msg interface{}) error {
	buf := self.buf
	defer buf.Clear()

	for {
		_, err := buf.ReadFrom(self.r)

		if err != nil {
			return err
		}

		for {
			if buf.Readable() > 0 {
				// 解码一个完整包
				if complete, data := self.doDecode(); complete {
					if nil != data {
						log.Infof("%s Receive a message <%s> from <%s>", MODULE_SERVER_BACKEND, string(data), self.backend.currentServer())

						m := make(map[string]interface{})
						d := json.NewDecoder(bytes.NewReader(data))
						d.UseNumber()
						err := d.Decode(&m)

						if err != nil {
							return err
						}

						self.backend.received(m)
					}

					log.Infof("%s has %d bytes to decode", MODULE_SERVER_BACKEND, buf.Readable())

					if buf.Readable() == 0 {
						buf.Clear()
					} else if buf.Readable() > 0 {
						continue
					} else {
						return nil
					}
				} else {
					// 数据不够，继续等待数据
					break
				}
			} else {
				// 数据不够，继续等待数据
				break
			}
		}
	}

	return nil
}

func (self myDecoder) doDecode() (bool, []byte) {
	readable := self.buf.Readable()

	log.Debugf("%s Received <%d> bytes, readed <%d>, write <%d>, capacity <%d>, writeable <%d>", MODULE_SERVER_NET, readable,
		self.buf.GetReaderIndex(), self.buf.GetWriteIndex(), self.buf.Capacity(), self.buf.Writeable())

	if readable < self.length {
		return false, nil
	}

	dataLength, _ := self.buf.PeekInt(0)

	log.Debugf("%s Data Length <%d> bytes", MODULE_SERVER_NET, dataLength)

	if 0 == dataLength {
		return true, nil
	}

	if readable < self.length+dataLength {
		return false, nil
	}

	self.buf.Skip(self.length)

	_, data, _ := self.buf.ReadBytes(dataLength)

	log.Debugf("%s Decoded data <%v>, <%s>", MODULE_SERVER_NET, data, string(data))

	if string(data) == "Ver1" {
		return true, nil
	}

	return true, data
}

func intToBytes(v int) []byte {
	ret := make([]byte, 4)
	ret[0] = byte(v >> 24)
	ret[1] = byte(v >> 16)
	ret[2] = byte(v >> 8)
	ret[3] = byte(v)
	return ret
}

