package store

type DB struct{}

func Open() *DB {
	return &DB{}
}
