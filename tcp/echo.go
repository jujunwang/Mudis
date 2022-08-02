package tcp

/**
 * 用于测试处理程序是否正常运行的echo处理程序
 */

import (
	"bufio"
	"context"
	"github.com/jujunwang/Mudis/lib/logger"
	"github.com/jujunwang/Mudis/lib/sync/atomic"
	"github.com/jujunwang/Mudis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

// EchoHandler 向客户端接收回复，用于测试
type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

// MakeEchoHandler 新建一个 EchoHandler
func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

// EchoClient 是 EchoHandler 的客户端, 用于测试
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// Close 关闭连接
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Conn.Close()
	return nil
}

// Handle 像客户端发送已收到的命令
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// 关闭处理程序拒绝新的连接
		_ = conn.Close()
	}

	client := &EchoClient{
		Conn: conn,
	}
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		// 可能发生:客户端EOF，客户端超时，处理器提前关闭
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("connection close")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

// Close 停止 echo 处理器
func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
