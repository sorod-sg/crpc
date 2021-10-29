package codec

import (
	"io"
)

type Header struct {
	ServiceMethod string //服务名和方法名
	Seq           uint64 //请求序号
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error         //写入header中
	ReadBody(interface{}) error       //写入body
	Write(*Header, interface{}) error //写入codec
}

type NewCodecHunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecHunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecHunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
