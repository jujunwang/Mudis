package cluster

import (
	"context"
	"errors"
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/lib/utils"
	"github.com/jujunwang/Mudis/resp/client"
	"github.com/jujunwang/Mudis/resp/reply"
	"strconv"
)

// 拿到与该节点的连接（先拿到与该节点的连接池，再从连接池里borrow一个连接）
func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	factory, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New("connection factory not found")
	}
	raw, err := factory.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	conn, ok := raw.(*client.Client)
	if !ok {
		return nil, errors.New("connection factory make wrong type")
	}
	return conn, nil
}

// 把连接换回连接池，防止连接耗尽
func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	connectionFactory, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection factory not found")
	}
	return connectionFactory.ReturnObject(context.Background(), peerClient)
}

// relay 将命令转发到节点
// 通过c.GetDBIndex()拿到数据库id，并转换到该数据库
// 不能调用 Prepare, Commit, execRollback等请求
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self {
		// to self db
		return cluster.db.Exec(c, args)
	}
	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}
	defer func() {
		_ = cluster.returnPeerClient(peer, peerClient)
	}()
	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex())))
	return peerClient.Send(args)
}

// broadcast 广播命令到集群中的所有节点
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	result := make(map[string]resp.Reply)
	for _, node := range cluster.nodes {
		reply := cluster.relay(node, c, args)
		result[node] = reply
	}
	return result
}
