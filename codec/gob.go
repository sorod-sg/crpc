package codec

import (
	"bufio"
	"encoding/gob"
	//"fmt"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn), //编码
		enc:  gob.NewEncoder(conn), //解码
	}
}
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
} //将head的编码值写入h中
func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
} //将body的编码写入body中

func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.conn.Close()
		}
	}()
	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
} //将h和body的json编码写入conn
func (c *GobCodec) Close() error {
	return c.conn.Close()
}
