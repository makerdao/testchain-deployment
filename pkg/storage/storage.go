package storage

type Storage interface {
	GetUpdate() bool
	GetRun() bool
}
