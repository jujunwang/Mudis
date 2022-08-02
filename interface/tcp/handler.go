package tcp

import (
	"context"
	"net"
)

// HandleFunc 表示命令对应的处理函数
type HandleFunc func(ctx context.Context, conn net.Conn)

// Handler 代表tcp上的应用处理函数
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
