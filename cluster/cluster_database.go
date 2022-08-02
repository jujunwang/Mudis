// Package cluster provides a server side cluster which is transparent to client. You can connect to any node in the cluster to access all data in the cluster
package cluster

import (
	"context"
	"fmt"
	pool "github.com/jolestar/go-commons-pool/v2"
	"github.com/jujunwang/Mudis/config"
	"github.com/jujunwang/Mudis/database"
	databaseface "github.com/jujunwang/Mudis/interface/database"
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/lib/consistenthash"
	"github.com/jujunwang/Mudis/lib/logger"
	"github.com/jujunwang/Mudis/resp/reply"
	"runtime/debug"
	"strings"
)

// ClusterDatabase 代表集群的一个节点
// 它保存部分数据并且协调其他节点完成服务
type ClusterDatabase struct {
	self string

	nodes          []string
	peerPicker     *consistenthash.NodeMap
	peerConnection map[string]*pool.ObjectPool
	db             databaseface.Database
}

// MakeClusterDatabase 创建并启动集群的一个节点
func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self: config.Properties.Self,

		db:             database.NewStandaloneDatabase(),
		peerPicker:     consistenthash.NewNodeMap(nil),
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	nodes = append(nodes, config.Properties.Self)
	cluster.peerPicker.AddNode(nodes...)
	ctx := context.Background()
	for _, peer := range config.Properties.Peers {
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	cluster.nodes = nodes
	return cluster
}

// CmdFunc 代表集群中一个与redis命令绑定的处理函数
// 集群版的command
type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdAndArgs [][]byte) resp.Reply

// Close 将停止集群中的当前节点
func (cluster *ClusterDatabase) Close() {
	cluster.db.Close()
}

var router = makeRouter()

// Exec 在集群上执行命令
// 这是解析后最先开始的工作
func (cluster *ClusterDatabase) Exec(c resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknownErrReply{}
		}
	}()
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "', or not supported in cluster mode")
	}
	result = cmdFunc(cluster, c, cmdLine)
	return
}

// AfterClientClose 做关闭后的清理工作
func (cluster *ClusterDatabase) AfterClientClose(c resp.Connection) {
	cluster.db.AfterClientClose(c)
}
