package reply

// PongReply is +PONG
type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")

// ToBytes 解析 redis.Reply
func (r *PongReply) ToBytes() []byte {
	return pongBytes
}

// OkReply -> +OK
type OkReply struct{}

var okBytes = []byte("+OK\r\n")

// ToBytes 解析 redis.Reply
func (r *OkReply) ToBytes() []byte {
	return okBytes
}

var theOkReply = new(OkReply)

// MakeOkReply 返回一个ok类型的reply
func MakeOkReply() *OkReply {
	return theOkReply
}

var nullBulkBytes = []byte("$-1\r\n")

// NullBulkReply 是一个空的字符串
type NullBulkReply struct{}

// ToBytes 解析 redis.Reply
func (r *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

// MakeNullBulkReply 新建一个新的 NullBulkReply
func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

var emptyMultiBulkBytes = []byte("*0\r\n")

// EmptyMultiBulkReply 是一个空的list
type EmptyMultiBulkReply struct{}

// ToBytes marshal redis.Reply
func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

// NoReply 对于像subscribe这样的命令什么也不回复
type NoReply struct{}

var noBytes = []byte("")

func (r *NoReply) ToBytes() []byte {
	return noBytes
}
