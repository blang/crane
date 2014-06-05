package store

// Interface for meta storage
type MetaStorage interface {
	ImageJSON(imageID string) (string, bool)
	TmpImageJSON(imageID string) (string, bool)
	SetTmpImageJSON(imageID string, json string) error
	Checksum(imageID string) (string, bool)
	TmpChecksum(imageID string) (string, bool)
	SetTmpChecksum(imageID string, checksum string) error
	Size(imageID string) (int64, bool)
	SetTmpSize(imageID string, size int64) error

	Ancestry(imageID string) ([]string, error)
	SetTmpAncestry(imageID string, parentImageID string) error
	Tags(namespace string, repository string) (map[string]string, bool)
	Tag(namespace string, repository string, tag string) (string, bool)
	SetTag(namespace string, repository string, imageID string, tag string) error
	SetImages(namespace string, repository string, images []string) error
	Images(namespace string, repository string) ([]string, error)
	// DeleteTag(namespace string, repository string, tag string) error
	// DeleteRepository(namespace string, repository string) error

	CommitTmpImage(imageID string) bool
	DiscardTmpImage(imageID string) bool
}
