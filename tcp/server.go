package tcp

/**
 * 一个tcp处理器
 */

import (
	"context"
	"fmt"
	"github.com/jujunwang/Mudis/interface/tcp"
	"github.com/jujunwang/Mudis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Config 存储tcp处理程序的配置
type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

// ListenAndServeWithSignal 绑定端口和处理请求，阻塞直到收到停止信号
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe 绑定端口和处理请求，阻塞直到关闭
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// listen signal
	go func() {
		<-closeChan
		logger.Info("shutting down...")
		_ = listener.Close() // listener.Accept() 会立即返回错误
		_ = handler.Close()  // close 连接
	}()

	// 监听端口
	defer func() {
		// 发生意外错误时关闭
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		// 传递给处理器
		logger.Info("accept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}
