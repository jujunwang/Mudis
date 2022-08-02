package database

import (
	"github.com/jujunwang/Mudis/interface/resp"
)

type CmdLine = [][]byte

// Database 是兼容Mudis的接口
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection)
	Close()
}

// DataEntity 存储指定 key 对应的数据, 包括 string, list, hash, set
type DataEntity struct {
	Data interface{}
}
