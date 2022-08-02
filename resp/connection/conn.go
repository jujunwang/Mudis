package connection

import (
	"bytes"
	"github.com/jujunwang/Mudis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

// Connection 代表一个与客户端的连接
type Connection struct {
	conn net.Conn
	// 用于等待请求处理结束
	waitingReply wait.Wait
	// 用于发送响应时加锁
	mu sync.Mutex
	// 切换DB
	selectedDB int
}

func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// RemoteAddr 返回远端地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close 与客户端断开连接
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// Write 通过TCP向客户端发送响应
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()

	_, err := c.conn.Write(b)
	return err
}

// GetDBIndex 返回选中的DB
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB 切换DB
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}

// FakeConn 假的 redis server
type FakeConn struct {
	Connection
	buf bytes.Buffer
}

func (c *FakeConn) Write(b []byte) error {
	c.buf.Write(b)
	return nil
}

// 清空缓存
func (c *FakeConn) Clean() {
	c.buf.Reset()
}

// Bytes 返回写入的key
func (c *FakeConn) Bytes() []byte {
	return c.buf.Bytes()
}
