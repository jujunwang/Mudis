package database

import (
	"github.com/jujunwang/Mudis/interface/database"
	"github.com/jujunwang/Mudis/interface/resp"
	"github.com/jujunwang/Mudis/lib/utils"
	"github.com/jujunwang/Mudis/resp/reply"
	"strconv"
	"strings"
)

func (db *DB) getAsString(key string) ([]byte, reply.ErrorReply) {
	entity, ok := db.GetEntity(key)
	if !ok {
		return nil, nil
	}
	bytes, ok := entity.Data.([]byte)
	if !ok {
		return nil, &reply.WrongTypeErrReply{}
	}
	return bytes, nil
}

// execGet 返回对应 key 的 string
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	bytes, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bytes == nil {
		return &reply.NullBulkReply{}
	}
	return reply.MakeBulkReply(bytes)
}

const (
	upsertPolicy = iota // default
	insertPolicy        // set nx
	updatePolicy        // set ex
)

// execSet 设置给定的 k v 键值对 和 ttl
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	policy := upsertPolicy
	// 解析
	if len(args) > 2 {
		for i := 2; i < len(args); i++ {
			arg := strings.ToUpper(string(args[i]))
			if arg == "NX" { // insert
				if policy == updatePolicy {
					return &reply.SyntaxErrReply{}
				}
				policy = insertPolicy
			} else if arg == "XX" { // update policy
				if policy == insertPolicy {
					return &reply.SyntaxErrReply{}
				}
				policy = updatePolicy
			} else {
				return &reply.SyntaxErrReply{}
			}
		}
	}

	entity := &database.DataEntity{
		Data: value,
	}

	var result int
	switch policy {
	case upsertPolicy:
		db.PutEntity(key, entity)
		result = 1
	case insertPolicy:
		result = db.PutIfAbsent(key, entity)
	case updatePolicy:
		result = db.PutIfExists(key, entity)
	}
	db.addAof(utils.ToCmdLine2("set", args...))
	if result > 0 {
		return &reply.OkReply{}
	}
	return &reply.NullBulkReply{}
}

// execSetNX 当给定 key 不存在时，才 set
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	result := db.PutIfAbsent(key, entity)
	db.addAof(utils.ToCmdLine2("setnx", args...))
	return reply.MakeIntReply(int64(result))
}

// execMSet 同时设置一个或多个 key-value 对
func execMSet(db *DB, args [][]byte) resp.Reply {
	if len(args)%2 != 0 {
		return reply.MakeSyntaxErrReply()
	}

	size := len(args) / 2
	keys := make([]string, size)
	values := make([][]byte, size)
	for i := 0; i < size; i++ {
		keys[i] = string(args[2*i])
		values[i] = args[2*i+1]
	}

	for i, key := range keys {
		value := values[i]
		db.PutEntity(key, &database.DataEntity{Data: value})
	}
	db.addAof(utils.ToCmdLine2("mset", args...))
	return &reply.OkReply{}
}

// execMGet 返回所有(一个或多个)给定 key 的值
func execMGet(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}

	result := make([][]byte, len(args))
	for i, key := range keys {
		bytes, err := db.getAsString(key)
		if err != nil {
			_, isWrongType := err.(*reply.WrongTypeErrReply)
			if isWrongType {
				result[i] = nil
				continue
			} else {
				return err
			}
		}
		result[i] = bytes // nil or []byte
	}

	return reply.MakeMultiBulkReply(result)
}

// execMSetNX 同时设置一个或多个 key-value 对，当且仅当所有给定 key 都不存在
func execMSetNX(db *DB, args [][]byte) resp.Reply {
	// parse args
	if len(args)%2 != 0 {
		return reply.MakeSyntaxErrReply()
	}
	size := len(args) / 2
	values := make([][]byte, size)
	keys := make([]string, size)
	for i := 0; i < size; i++ {
		keys[i] = string(args[2*i])
		values[i] = args[2*i+1]
	}

	for _, key := range keys {
		_, exists := db.GetEntity(key)
		if exists {
			return reply.MakeIntReply(0)
		}
	}

	for i, key := range keys {
		value := values[i]
		db.PutEntity(key, &database.DataEntity{Data: value})
	}
	db.addAof(utils.ToCmdLine2("msetnx", args...))
	return reply.MakeIntReply(1)
}

// execGetSet 设置字符串类型键的值并返回其原来的值
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]

	old, err := db.getAsString(key)
	if err != nil {
		return err
	}
	db.PutEntity(key, &database.DataEntity{Data: value})
	if old == nil {
		return new(reply.NullBulkReply)
	}
	db.addAof(utils.ToCmdLine2("getset", args...))
	return reply.MakeBulkReply(old)
}

// execIncr 对指定 key 的 value 加一
func execIncr(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	bytes, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bytes != nil {
		val, err := strconv.ParseInt(string(bytes), 10, 64)
		if err != nil {
			return reply.MakeErrReply("ERR value is not an integer or out of range")
		}
		db.PutEntity(key, &database.DataEntity{
			Data: []byte(strconv.FormatInt(val+1, 10)),
		})
		db.addAof(utils.ToCmdLine2("incr", args...))
		return reply.MakeIntReply(val + 1)
	}
	db.PutEntity(key, &database.DataEntity{
		Data: []byte("1"),
	})
	db.addAof(utils.ToCmdLine2("incr", args...))
	return reply.MakeIntReply(1)
}

// execIncrBy 对指定 key 的 value 加给定值
func execIncrBy(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	rawDelta := string(args[1])
	delta, err := strconv.ParseInt(rawDelta, 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}

	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes != nil {
		// existed value
		val, err := strconv.ParseInt(string(bytes), 10, 64)
		if err != nil {
			return reply.MakeErrReply("ERR value is not an integer or out of range")
		}
		db.PutEntity(key, &database.DataEntity{
			Data: []byte(strconv.FormatInt(val+delta, 10)),
		})
		db.addAof(utils.ToCmdLine2("incrby", args...))
		return reply.MakeIntReply(val + delta)
	}
	db.PutEntity(key, &database.DataEntity{
		Data: args[1],
	})
	db.addAof(utils.ToCmdLine2("incrby", args...))
	return reply.MakeIntReply(delta)
}

// execDecr 对指定 key 的 value 减一
func execDecr(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes != nil {
		val, err := strconv.ParseInt(string(bytes), 10, 64)
		if err != nil {
			return reply.MakeErrReply("ERR value is not an integer or out of range")
		}
		db.PutEntity(key, &database.DataEntity{
			Data: []byte(strconv.FormatInt(val-1, 10)),
		})
		db.addAof(utils.ToCmdLine2("decr", args...))
		return reply.MakeIntReply(val - 1)
	}
	entity := &database.DataEntity{
		Data: []byte("-1"),
	}
	db.PutEntity(key, entity)
	db.addAof(utils.ToCmdLine2("decr", args...))
	return reply.MakeIntReply(-1)
}

// execDecrBy 对指定 key 的 value 减指定值
func execDecrBy(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	rawDelta := string(args[1])
	delta, err := strconv.ParseInt(rawDelta, 10, 64)
	if err != nil {
		return reply.MakeErrReply("ERR value is not an integer or out of range")
	}

	bytes, errReply := db.getAsString(key)
	if errReply != nil {
		return errReply
	}
	if bytes != nil {
		val, err := strconv.ParseInt(string(bytes), 10, 64)
		if err != nil {
			return reply.MakeErrReply("ERR value is not an integer or out of range")
		}
		db.PutEntity(key, &database.DataEntity{
			Data: []byte(strconv.FormatInt(val-delta, 10)),
		})
		db.addAof(utils.ToCmdLine2("decrby", args...))
		return reply.MakeIntReply(val - delta)
	}
	valueStr := strconv.FormatInt(-delta, 10)
	db.PutEntity(key, &database.DataEntity{
		Data: []byte(valueStr),
	})
	db.addAof(utils.ToCmdLine2("decrby", args...))
	return reply.MakeIntReply(-delta)
}

// execStrLen 返回指定 key 的 value 的长度
func execStrLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	bytes, err := db.getAsString(key)
	if err != nil {
		return err
	}
	if bytes == nil {
		return reply.MakeIntReply(0)
	}
	return reply.MakeIntReply(int64(len(bytes)))
}

// execAppend 对指定 key 进行修改
func execAppend(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	bytes, err := db.getAsString(key)
	if err != nil {
		return err
	}
	bytes = append(bytes, args[1]...)
	db.PutEntity(key, &database.DataEntity{
		Data: bytes,
	})
	db.addAof(utils.ToCmdLine2("append", args...))
	return reply.MakeIntReply(int64(len(bytes)))
}

// execSetRange 覆盖按键存储的字符串的一部分，从指定的偏移量开始。
// 如果偏移量大于字符串 key 的当前长度，则字符串被填充为零字节
func execSetRange(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	offset, errNative := strconv.ParseInt(string(args[1]), 10, 64)
	if errNative != nil {
		return reply.MakeErrReply(errNative.Error())
	}
	value := args[2]
	bytes, err := db.getAsString(key)
	if err != nil {
		return err
	}
	bytesLen := int64(len(bytes))
	if bytesLen < offset {
		diff := offset - bytesLen
		diffArray := make([]byte, diff)
		bytes = append(bytes, diffArray...)
		bytesLen = int64(len(bytes))
	}
	for i := 0; i < len(value); i++ {
		idx := offset + int64(i)
		if idx >= bytesLen {
			bytes = append(bytes, value[i])
		} else {
			bytes[idx] = value[i]
		}
	}
	db.PutEntity(key, &database.DataEntity{
		Data: bytes,
	})
	db.addAof(utils.ToCmdLine2("setRange", args...))
	return reply.MakeIntReply(int64(len(bytes)))
}

func execGetRange(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	startIdx, errNative := strconv.ParseInt(string(args[1]), 10, 64)
	if errNative != nil {
		return reply.MakeErrReply(errNative.Error())
	}
	endIdx, errNative := strconv.ParseInt(string(args[2]), 10, 64)
	if errNative != nil {
		return reply.MakeErrReply(errNative.Error())
	}

	bytes, err := db.getAsString(key)
	if err != nil {
		return err
	}

	if bytes == nil {
		return reply.MakeNullBulkReply()
	}

	bytesLen := int64(len(bytes))
	if startIdx < -1*bytesLen {
		return &reply.NullBulkReply{}
	} else if startIdx < 0 {
		startIdx = bytesLen + startIdx
	} else if startIdx >= bytesLen {
		return &reply.NullBulkReply{}
	}
	if endIdx < -1*bytesLen {
		return &reply.NullBulkReply{}
	} else if endIdx < 0 {
		endIdx = bytesLen + endIdx + 1
	} else if endIdx < bytesLen {
		endIdx = endIdx + 1
	} else {
		endIdx = bytesLen
	}
	if startIdx > endIdx {
		return reply.MakeNullBulkReply()
	}

	return reply.MakeBulkReply(bytes[startIdx:endIdx])
}

func init() {
	RegisterCommand("Set", execSet, -3)
	RegisterCommand("SetNx", execSetNX, 3)
	RegisterCommand("MSet", execMSet, -3)
	RegisterCommand("MGet", execMGet, -2)
	RegisterCommand("MSetNX", execMSetNX, -3)
	RegisterCommand("Get", execGet, 2)
	RegisterCommand("GetSet", execGetSet, 3)
	RegisterCommand("Incr", execIncr, 2)
	RegisterCommand("IncrBy", execIncrBy, 3)
	RegisterCommand("Decr", execDecr, 2)
	RegisterCommand("DecrBy", execDecrBy, 3)
	RegisterCommand("StrLen", execStrLen, 2)
	RegisterCommand("Append", execAppend, 3)
	RegisterCommand("SetRange", execSetRange, 4)
	RegisterCommand("GetRange", execGetRange, 4)
}
