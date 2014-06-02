package main

// Interface for meta storage
type MetaStorage interface {
	SetImageJSON(imageID string, json string) error
	ImageJSON(imageID string) (string, error)
	Ancestry(imageID string) (string, error)
	Tags(namespace string, repository string) (string, error)
	Tag(namespace string, repository string, tag string) (string, error)
	SetTag(namespace string, repository string, tag string) error
	DeleteTag(namespace string, repository string, tag string) error
	DeleteRepository(namespace string, repository string) error
}
