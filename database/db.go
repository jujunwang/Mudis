// database 包是一个具有redis兼容接口的内存数据库
package database

import (
	"github.com/jujunwang/Mudis/datastruct/dict"
	"github.com/jujunwang/Mudis/interface/database"
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/resp/reply"
	"strings"
)

const (
	dataDictSize = 1 << 16
)

// DB 存储数据、执行用户的命令
type DB struct {
	index int
	// key -> DataEntity
	data   dict.Dict
	addAof func(CmdLine)
}

// ExecFunc 是命令对应函数的接口
// args 不包含 cmd 列，例如：set a b ——> a b
type ExecFunc func(db *DB, args [][]byte) resp.Reply

// CmdLine 代表命令行
type CmdLine = [][]byte

// makeDB 创建 DB 实例
func makeDB() *DB {
	db := &DB{
		//data:   dict.MakeSyncDict(),
		// 换用分段锁实现hashmap
		data:   dict.MakeConcurrent(dataDictSize),
		addAof: func(line CmdLine) {},
	}
	return db
}

// Exec 在单机数据库中执行命令
func (db *DB) Exec(c resp.Connection, cmdLine [][]byte) resp.Reply {

	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "'")
	}
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.executor
	return fun(db, cmdLine[1:])
}

func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	}
	return argNum >= -arity
}

/* ---- 数据存取 ----- */

//Entity 是db这一层操作(存取)的对象，指代 Mudis 的各种数据类型

// GetEntity 返回绑定到给定键的DataEntity
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {

	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

// PutEntity 向 DB 中写入DataEntity
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists 编辑已存在的数据库实体
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

// PutIfAbsent 当且仅当key不存在时插入一个 DataEntity
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove 从数据库中删除指定key
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// 一次性删除多个 key
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

// Flush 清空 database
func (db *DB) Flush() {
	db.data.Clear()
}
