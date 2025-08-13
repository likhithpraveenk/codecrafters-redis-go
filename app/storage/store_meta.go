package store

func GetType(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	it, exists := store[key]
	if !exists {
		return "none"
	}
	switch it.typ {
	case TypeString:
		return "string"
	case TypeList:
		return "list"
	case TypeStream:
		return "stream"
	case TypeHash:
		return "hash"
	case TypeSet:
		return "set"
	case TypeZSet:
		return "zset"
	default:
		return "none"
	}
}
