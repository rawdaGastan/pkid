package internal

// PkidStore an interface for pkid db store
type PkidStore interface {
	setConn(string) error
	migrate() error
	get(string) (string, error)
	set(string, string) error
	update(string, string) error
	delete(string) error
	list() ([]string, error)
}
