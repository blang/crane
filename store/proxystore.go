package store

type ProxyStore struct {
	MetaStorage
	FileStorage
}

func NewProxyStore(meta MetaStorage, file FileStorage) *ProxyStore {
	return &ProxyStore{
		MetaStorage: meta,
		FileStorage: file,
	}
}
