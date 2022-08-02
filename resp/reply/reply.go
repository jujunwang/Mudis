package reply

import (
	"bytes"
	"github.com/jujunwang/Mudis/interface/resp"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")

	// CRLF 是redis序列化协议的行分隔符
	CRLF = "\r\n"
)

/* ---- 多行 Reply ---- */

// BulkReply存储一个二进制安全的字符串
type BulkReply struct {
	Arg []byte
}

// MakeBulkReply 返回一个多行 reply
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

// ToBytes 解析 redis.Reply
func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

/* ---- Multi Bulk Reply ---- */

// MultiBulkReply 存储一个字符串列表
type MultiBulkReply struct {
	Args [][]byte
}

// MakeMultiBulkReply 新建一个 MultiBulkReply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

// ToBytes 解析 redis.Reply
func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

/* ---- Status Reply ---- */

// StatusReply 存储一个string来表示状态
type StatusReply struct {
	Status string
}

// MakeStatusReply 返回 StatusReply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

/* ---- Int Reply ---- */

// IntReply 存储一个 int64 类型的数字
type IntReply struct {
	Code int64
}

// MakeIntReply 返回一个int类型的reply
func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// ToBytes 解析 redis.Reply
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

/* ---- Error Reply ---- */

// ErrorReply error 类型的 reply
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// StandardErrReply 表示处理器错误
type StandardErrReply struct {
	Status string
}

// ToBytes 解析 redis.Reply
func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

// MakeErrReply 返回 StandardErrReply
func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// IsErrorReply 如果给定的reply是错误，返回true
func IsErrorReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
