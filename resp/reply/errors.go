package reply

// UnknownErrReply 表示 UnknownErr
type UnknownErrReply struct{}

var unknownErrBytes = []byte("-Err unknown\r\n")

// ToBytes 解析 redis.Reply
func (r *UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

func (r *UnknownErrReply) Error() string {
	return "Err unknown"
}

// ArgNumErrReply 表示命令的参数数目错误
type ArgNumErrReply struct {
	Cmd string
}

// ToBytes 解析 redis.Reply
func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}

func (r *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + r.Cmd + "' command"
}

// MakeArgNumErrReply 表示命令的参数数目错误
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{
		Cmd: cmd,
	}
}

// SyntaxErrReply 表示遇到错误的声明
type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-Err syntax error\r\n")
var theSyntaxErrReply = &SyntaxErrReply{}

// MakeSyntaxErrReply 新建一个 syntax error
func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

// ToBytes 解析 redis.Reply
func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func (r *SyntaxErrReply) Error() string {
	return "Err syntax error"
}

// WrongTypeErrReply 表示对持有错误类型值的键的操作
type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

// ToBytes 解析 redis.Reply
func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}

// ProtocolErr

// ProtocolErrReply 表示在解析请求期间遇到错误类型的数据
type ProtocolErrReply struct {
	Msg string
}

// ToBytes 解析 redis.Reply
func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol error: '" + r.Msg + "'\r\n")
}

func (r *ProtocolErrReply) Error() string {
	return "ERR Protocol error: '" + r.Msg
}
