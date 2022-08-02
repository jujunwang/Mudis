package handler

/*
 * tcp.RespHandler 实现了 RESP 协议的接口
 */

import (
	"context"
	"github.com/jujunwang/Mudis/cluster"
	"github.com/jujunwang/Mudis/config"
	"github.com/jujunwang/Mudis/database"
	databaseface "github.com/jujunwang/Mudis/interface/database"
	"github.com/jujunwang/Mudis/lib/logger"
	"github.com/jujunwang/Mudis/lib/sync/atomic"
	"github.com/jujunwang/Mudis/resp/connection"
	"github.com/jujunwang/Mudis/resp/parser"
	"github.com/jujunwang/Mudis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

// RespHandler 实现了 tcp.Handler 并且作为一个 redis 服务器提供服务
type RespHandler struct {
	activeConn sync.Map // *client -> placeholder
	db         databaseface.Database
	closing    atomic.Boolean // refusing new client and new request
}

// MakeHandler 新建一个 RespHandler 实例
func MakeHandler() *RespHandler {
	var db databaseface.Database
	if config.Properties.Self != "" &&
		len(config.Properties.Peers) > 0 {
		db = cluster.MakeClusterDatabase()
	} else {
		db = database.NewStandaloneDatabase()
	}

	return &RespHandler{
		db: db,
	}
}

func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Handle 接收并执行redis命令
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// 关闭处理程序拒绝新的连接
		_ = conn.Close()
	}

	client := connection.NewConn(conn)
	h.activeConn.Store(client, 1)

	ch := parser.ParseStream(conn)
	for payload := range ch {
		//Err
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close 停止处理器
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}
