package cache

type Value interface {
	Len() int
}

type BaseCache interface {
	Get(key string) (Value, bool)
	Add(key string, value Value)
}
