package db

type KeyDB interface {
	Close() error
	Migrate() error
	AddKey(key string) error
	GetKeys() ([]string, error)
	DeleteKey(key string) error
	TruncateKeys() error
}
