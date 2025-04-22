package hash

type Hasher interface {
	Hash(pwd string) (string, error)
	Compare(hash, pwd string) error
}
