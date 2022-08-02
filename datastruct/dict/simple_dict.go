package dict

// SimpleDict 封装一个map，它不是线程安全的
type SimpleDict struct {
	m map[string]interface{}
}

// MakeSimple 新建一个map
func MakeSimple() *SimpleDict {
	return &SimpleDict{
		m: make(map[string]interface{}),
	}
}

// Get 返回绑定值以及该键是否存在
func (dict *SimpleDict) Get(key string) (val interface{}, exists bool) {
	val, ok := dict.m[key]
	return val, ok
}

// Len 返回 dict 的长度
func (dict *SimpleDict) Len() int {
	if dict.m == nil {
		panic("m is nil")
	}
	return len(dict.m)
}

// Put 将键值放入字典并返回新插入键值对的个数
func (dict *SimpleDict) Put(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	dict.m[key] = val
	if existed {
		return 0
	}
	return 1
}

// PutIfAbsent 如果键不存在则返回值，并返回更新的键值对的个数
func (dict *SimpleDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	if existed {
		return 0
	}
	dict.m[key] = val
	return 1
}

// PutIfExists 如果键存在，则写入值，并返回插入键值对的个数
func (dict *SimpleDict) PutIfExists(key string, val interface{}) (result int) {
	_, existed := dict.m[key]
	if existed {
		dict.m[key] = val
		return 1
	}
	return 0
}

// Remove 删除键并返回已删除键值对的个数
func (dict *SimpleDict) Remove(key string) (result int) {
	_, existed := dict.m[key]
	delete(dict.m, key)
	if existed {
		return 1
	}
	return 0
}

// Keys 返回dict中的所有key
func (dict *SimpleDict) Keys() []string {
	result := make([]string, len(dict.m))
	i := 0
	for k := range dict.m {
		result[i] = k
	}
	return result
}

// ForEach 遍历dict
func (dict *SimpleDict) ForEach(consumer Consumer) {
	for k, v := range dict.m {
		if !consumer(k, v) {
			break
		}
	}
}

// RandomKeys 随机返回给定数字的键，可能包含重复的键
func (dict *SimpleDict) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		for k := range dict.m {
			result[i] = k
			break
		}
	}
	return result
}

// RandomDistinctKeys 随机返回给定数字的键，不会包含重复的键
func (dict *SimpleDict) RandomDistinctKeys(limit int) []string {
	size := limit
	if size > len(dict.m) {
		size = len(dict.m)
	}
	result := make([]string, size)
	i := 0
	for k := range dict.m {
		if i == limit {
			break
		}
		result[i] = k
		i++
	}
	return result
}

// Clear 清空dict中的key
func (dict *SimpleDict) Clear() {
	*dict = *MakeSimple()
}
