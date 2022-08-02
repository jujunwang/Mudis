package database

import (
	"fmt"
	"github.com/jujunwang/Mudis/aof"
	"github.com/jujunwang/Mudis/config"
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/lib/logger"
	"github.com/jujunwang/Mudis/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

// StandaloneDatabase 是多个单机数据库
type StandaloneDatabase struct {
	dbSet []*DB
	// aof 持久化处理器
	aofHandler *aof.AofHandler
}

// NewStandaloneDatabase 新建一个 redis 实例,
func NewStandaloneDatabase() *StandaloneDatabase {
	mdb := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	mdb.dbSet = make([]*DB, config.Properties.Databases)
	for i := range mdb.dbSet {
		singleDB := makeDB()
		singleDB.index = i
		mdb.dbSet[i] = singleDB
	}
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAOFHandler(mdb)
		if err != nil {
			panic(err)
		}
		mdb.aofHandler = aofHandler
		for _, db := range mdb.dbSet {
			// avoid closure
			singleDB := db
			singleDB.addAof = func(line CmdLine) {
				mdb.aofHandler.AddAof(singleDB.index, line)
			}
		}
	}
	return mdb
}

// Exec 执行命令
// 参数'cmdLine'包含命令及其参数，例如:"set key value"
func (mdb *StandaloneDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {

	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()

	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "select" {
		if len(cmdLine) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(c, mdb, cmdLine[1:])
	}
	// 普通命令
	dbIndex := c.GetDBIndex()
	if dbIndex >= len(mdb.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	selectedDB := mdb.dbSet[dbIndex]
	return selectedDB.Exec(c, cmdLine)
}

// Close 优雅关闭数据库
func (mdb *StandaloneDatabase) Close() {

}

func (mdb *StandaloneDatabase) AfterClientClose(c resp.Connection) {

}

func execSelect(c resp.Connection, mdb *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(mdb.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
