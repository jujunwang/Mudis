package database

import (
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/resp/reply"
)

// 向服务器发送PING
func Ping(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return &reply.PongReply{}
	} else if len(args) == 1 {
		return reply.MakeStatusReply(string(args[0]))
	} else {
		return reply.MakeErrReply("ERR wrong number of arguments for 'ping' command")
	}
}

func init() {
	RegisterCommand("ping", Ping, -1)
}
