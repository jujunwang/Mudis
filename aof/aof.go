package aof

import (
	"github.com/jujunwang/Mudis/config"
	databaseface "github.com/jujunwang/Mudis/interface/database"
	"github.com/jujunwang/Mudis/lib/logger"
	"github.com/jujunwang/Mudis/lib/utils"
	"github.com/jujunwang/Mudis/resp/connection"
	"github.com/jujunwang/Mudis/resp/parser"
	"github.com/jujunwang/Mudis/resp/reply"
	"io"
	"os"
	"strconv"
	"sync"
)

// CmdLine 代表命令
type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16
)

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

// AofHandler 从channel中获取数据，向AOF文件中写入数据
type AofHandler struct {
	db          databaseface.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	// 当AOF 执行完毕时，AOF goroutine会通过这个管道向主程序发送消息
	aofFinished chan struct{}
	// 解析 AOF 文件的时候，需要加锁
	pausingAof sync.RWMutex
	// 记录上一条指令工作在哪个db，以此来判断需不需要select
	currentDB int
}

// NewAOFHandler 新建一个新的 aof.AofHandler
func NewAOFHandler(db databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.db = db
	handler.LoadAof(0)
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize)
	handler.aofFinished = make(chan struct{})
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// AddAof 将命令塞到 channel 里
func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// handleAof 从 channel 里读命令并且将命令写入 AOF 文件
func (handler *AofHandler) handleAof() {
	handler.currentDB = 0
	for p := range handler.aofChan {
		//防止其他 goroutine 暂停 aof 过程
		handler.pausingAof.RLock()
		if p.dbIndex != handler.currentDB {
			// 切换db
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue
			}
			handler.currentDB = p.dbIndex
		}
		// 将用户的指令变成 RESP 协议的格式
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
		handler.pausingAof.RUnlock()
	}
	handler.aofFinished <- struct{}{}
}

// LoadAof 读 aof 文件
func (handler *AofHandler) LoadAof(maxBytes int) {
	// 删除 aofChan 防止再次重写
	aofChan := handler.aofChan
	handler.aofChan = nil
	defer func(aofChan chan *payload) {
		handler.aofChan = aofChan
	}(aofChan)

	file, err := os.Open(handler.aofFilename)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return
		}
		logger.Warn(err)
		return
	}
	defer file.Close()

	var reader io.Reader
	if maxBytes > 0 {
		reader = io.LimitReader(file, int64(maxBytes))
	} else {
		reader = file
	}
	// 解析AOF文件
	ch := parser.ParseStream(reader)
	//用来记录工作在哪个db，以判断用不用切换db
	fakeConn := &connection.FakeConn{}
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		ret := handler.db.Exec(fakeConn, r.Args)
		if reply.IsErrorReply(ret) {
			logger.Error("exec err", err)
		}
	}
}

// Close 优雅地停止一个持久化过程
func (handler *AofHandler) Close() {
	if handler.aofFile != nil {
		close(handler.aofChan)
		//等待AOF过程结束
		<-handler.aofFinished
		err := handler.aofFile.Close()
		if err != nil {
			logger.Warn(err)
		}
	}
}
