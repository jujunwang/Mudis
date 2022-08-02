package resp

// Reply 是一个符合RESP协议的消息接口
type Reply interface {
	ToBytes() []byte
}
