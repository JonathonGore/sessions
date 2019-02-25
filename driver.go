package sessions

type StorageDriver interface {
	InsertSession(s Session) error
	GetSession(sid string) (Session, error)
	DeleteSession(sid string) error
}
