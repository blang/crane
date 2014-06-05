package store

import (
	"errors"
)

type Repository struct {
	Images []string          // image ids
	Tags   map[string]string //tag -> imageid
}

func NewRepository() *Repository {
	return &Repository{
		Tags: make(map[string]string),
	}
}

type MemMetaStorage struct {
	imageTmpJsonMap     map[string]string
	imageTmpChecksumMap map[string]string
	imageTmpSizeMap     map[string]int64
	imageJsonMap        map[string]string
	imageChecksumMap    map[string]string
	imageSizeMap        map[string]int64
	imageAncestryMap    map[string]string
	imageTmpAncestryMap map[string]string
	repositoryMap       map[string]*Repository
}

func NewMemMetaStorage() *MemMetaStorage {
	return &MemMetaStorage{
		imageTmpJsonMap:     make(map[string]string),
		imageTmpChecksumMap: make(map[string]string),
		imageTmpSizeMap:     make(map[string]int64),
		imageJsonMap:        make(map[string]string),
		imageChecksumMap:    make(map[string]string),
		imageSizeMap:        make(map[string]int64),
		imageAncestryMap:    make(map[string]string),
		imageTmpAncestryMap: make(map[string]string),
		repositoryMap:       make(map[string]*Repository),
	}
}

func (m *MemMetaStorage) CommitTmpImage(imageID string) bool {
	json, found := m.imageTmpJsonMap[imageID]
	if !found {
		return false
	}
	checksum, found := m.imageTmpChecksumMap[imageID]
	if !found {
		return false
	}
	size, found := m.imageTmpSizeMap[imageID]
	if !found {
		return false
	}
	ancestry, ancestryFound := m.imageTmpAncestryMap[imageID]

	// Insert
	m.imageJsonMap[imageID] = json
	m.imageChecksumMap[imageID] = checksum
	m.imageSizeMap[imageID] = size
	if ancestryFound {
		m.imageAncestryMap[imageID] = ancestry
	}

	// Remove tmp
	delete(m.imageTmpJsonMap, imageID)
	delete(m.imageTmpChecksumMap, imageID)
	delete(m.imageTmpSizeMap, imageID)
	if ancestryFound {
		delete(m.imageTmpAncestryMap, imageID)
	}
	return true
}
func (m *MemMetaStorage) DiscardTmpImage(imageID string) bool {
	// Remove tmp
	delete(m.imageTmpJsonMap, imageID)
	delete(m.imageTmpChecksumMap, imageID)
	delete(m.imageTmpSizeMap, imageID)
	delete(m.imageTmpAncestryMap, imageID)
	return true
}

func (m *MemMetaStorage) SetTmpImageJSON(imageID string, json string) error {
	m.imageTmpJsonMap[imageID] = json
	return nil
}
func (m *MemMetaStorage) ImageJSON(imageID string) (string, bool) {
	json, found := m.imageJsonMap[imageID]
	return json, found
}
func (m *MemMetaStorage) TmpImageJSON(imageID string) (string, bool) {
	json, found := m.imageTmpJsonMap[imageID]
	return json, found
}

func (m *MemMetaStorage) SetTmpChecksum(imageID string, checksum string) error {
	m.imageTmpChecksumMap[imageID] = checksum
	return nil
}

func (m *MemMetaStorage) TmpChecksum(imageID string) (string, bool) {
	chs, found := m.imageTmpChecksumMap[imageID]
	return chs, found
}
func (m *MemMetaStorage) Checksum(imageID string) (string, bool) {
	chs, found := m.imageChecksumMap[imageID]
	return chs, found
}

func (m *MemMetaStorage) SetTmpSize(imageID string, size int64) error {
	m.imageTmpSizeMap[imageID] = size
	return nil
}
func (m *MemMetaStorage) Size(imageID string) (int64, bool) {
	size, found := m.imageSizeMap[imageID]
	return size, found
}

func (m *MemMetaStorage) Tags(namespace string, repository string) (map[string]string, bool) {
	repo, found := m.repositoryMap[namespace+"/"+repository]
	if !found {
		return nil, false
	}
	if len(repo.Tags) > 0 {
		return repo.Tags, true
	} else {
		return nil, false
	}
}

func (m *MemMetaStorage) Tag(namespace string, repository string, tag string) (string, bool) {
	repo, found := m.repositoryMap[namespace+"/"+repository]
	if !found {
		return "", false
	}
	imageID, found := repo.Tags[tag]
	return imageID, found
}

func (m *MemMetaStorage) SetTag(namespace string, repository string, imageID string, tag string) error {
	repo, found := m.repositoryMap[namespace+"/"+repository]
	if !found {
		repo = NewRepository()
		m.repositoryMap[namespace+"/"+repository] = repo
	}
	repo.Tags[tag] = imageID
	return nil
}

func (m *MemMetaStorage) SetImages(namespace string, repository string, images []string) error {
	repo, found := m.repositoryMap[namespace+"/"+repository]
	if !found {
		repo = NewRepository()
		m.repositoryMap[namespace+"/"+repository] = repo
	}
	repo.Images = images
	return nil
}
func (m *MemMetaStorage) Images(namespace string, repository string) ([]string, error) {
	repo, found := m.repositoryMap[namespace+"/"+repository]
	if !found {
		return nil, errors.New("Repository not found")
	}
	return repo.Images, nil
}

func (m *MemMetaStorage) Ancestry(imageID string) ([]string, error) {
	var ancestryArr []string
	for {
		ancestryArr = append(ancestryArr, imageID)
		if parentID, found := m.imageAncestryMap[imageID]; found {
			imageID = parentID
		} else {
			break
		}
	}
	return ancestryArr, nil
}

func (m *MemMetaStorage) SetTmpAncestry(imageID string, parentImageID string) error {
	m.imageTmpAncestryMap[imageID] = parentImageID
	return nil
}
