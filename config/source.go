package config

// KeyValue is config key value.
type KeyValue struct {
	Key    string
	Value  []byte
	Format string
}

// Source is config source.
// 配置源
type Source interface {
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
}

// Watcher watches a source for changes.
// 监听配置源变化
type Watcher interface {
	Next() ([]*KeyValue, error)
	Stop() error
}
