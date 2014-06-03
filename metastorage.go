package main

// Interface for meta storage
type MetaStorage interface {
	SetImageJSON(imageID string, json string) error
	ImageJSON(imageID string) (string, bool)
	SetChecksum(imageID string, checksum string) error
	Checksum(imageID string) (string, bool)
	SetSize(imageID string, size int64) error
	Size(imageID string) (int64, bool)
	Ancestry(imageID string) ([]string, error)
	SetAncestry(imageID string, parentImageID string) error
	Tags(namespace string, repository string) (map[string]string, bool)
	Tag(namespace string, repository string, tag string) (string, bool)
	SetTag(namespace string, repository string, imageID string, tag string) error
	SetImages(namespace string, repository string, images []string) error
	Images(namespace string, repository string) ([]string, error)
	// DeleteTag(namespace string, repository string, tag string) error
	// DeleteRepository(namespace string, repository string) error

}
