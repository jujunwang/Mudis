package database

import (
	"strings"
)

var cmdTable = make(map[string]*command)

type command struct {
	executor ExecFunc
	// 合法的命令args的长度，当arity < 0代表 args 的长度 >= arity
	arity int
}

// RegisterCommand 注册一个新命令
// arity 表示合法的cmdArgs长度, arity < 0 意味着 len(args) >= -arity. 例如: `get` 是 2, `mget` 是 -2
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		executor: executor,
		arity:    arity,
	}
}
