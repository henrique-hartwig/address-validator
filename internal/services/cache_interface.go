package services

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Flush()
	ItemCount() int
	Close() error
}
