package resp

// Connection 代表一个与客户端的连接
type Connection interface {
	Write([]byte) error
	// used for multi database
	GetDBIndex() int
	SelectDB(int)
}
