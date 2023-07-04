// package store is for pkid storage
package store

// PkidStore an interface for pkid db store
type PkidStore interface {
	SetConn(string) error
	Migrate() error
	Get(string) (string, error)
	Set(string, string) error
	Update(string, string) error
	Delete(string) error
	List() ([]string, error)
}
