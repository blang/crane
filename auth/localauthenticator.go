package auth

import (
	"crypto/rand"
	"encoding/base64"
)

type tokenPerm struct {
	namespace  string
	repository string
	images     map[string]bool
	mode       Mode
}

func newAuthToken(namespace, repository string, images []string, mode Mode) (*tokenPerm, string) {
	imap := make(map[string]bool)
	for _, i := range images {
		imap[i] = true
	}
	token := createRandomToken()
	return &tokenPerm{
		namespace:  namespace,
		repository: repository,
		images:     imap,
		mode:       mode,
	}, token
}

func createRandomToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

type LocalAuthenticator struct {
	tokenMap map[string]*tokenPerm
}

//TODO: Cleanup map
func NewLocalAuthenticator() *LocalAuthenticator {
	return &LocalAuthenticator{
		tokenMap: make(map[string]*tokenPerm),
	}
}

// Authenticate all users, because this is already done by proxy
func (l *LocalAuthenticator) Authenticate(user string, pass string) bool {
	return true
}

// Grant access to users namespace only
func (l *LocalAuthenticator) Authorize(user, pass, namespace, repository string, imageIDs []string, mode Mode) (string, bool) {
	if user != namespace {
		return "", false
	}
	tokenPerm, token := newAuthToken(namespace, repository, imageIDs, mode)
	l.tokenMap[token] = tokenPerm
	return token, true
}

func (l *LocalAuthenticator) HasPermPushImage(token string, imageID string) bool {
	perm, found := l.tokenMap[token]
	if !found {
		return false
	}
	if perm.mode == O_RDONLY {
		return false
	}
	if _, found = perm.images[imageID]; !found {
		return false
	}
	return true
}

func (l *LocalAuthenticator) HasPermPullImage(token string, imageID string) bool {
	perm, found := l.tokenMap[token]
	if !found {
		return false
	}
	if perm.mode != O_RDONLY {
		return false
	}
	if _, found = perm.images[imageID]; !found {
		return false
	}
	return true
}

func (l *LocalAuthenticator) HasPermPushTag(token string, namespace, repository string, imageID string, tag string) bool {
	perm, found := l.tokenMap[token]
	if !found {
		return false
	}
	if perm.mode == O_RDONLY {
		return false
	}
	if _, found = perm.images[imageID]; !found {
		return false
	}
	if !(perm.namespace == namespace && perm.repository == repository) {
		return false
	}
	return true
}

func (l *LocalAuthenticator) HasPermPullTag(token string, namespace, repository string, tag string) bool {
	perm, found := l.tokenMap[token]
	if !found {
		return false
	}
	if perm.mode != O_RDONLY {
		return false
	}
	if !(perm.namespace == namespace && perm.repository == repository) {
		return false
	}
	return true
}

func (l *LocalAuthenticator) HasPermPullTags(token string, namespace, repository string) bool {
	perm, found := l.tokenMap[token]
	if !found {
		return false
	}
	if perm.mode != O_RDONLY {
		return false
	}
	if !(perm.namespace == namespace && perm.repository == repository) {
		return false
	}
	return true
}

func (l *LocalAuthenticator) HasPermPushChecksums(token string, namespace, repository string) bool {
	perm, found := l.tokenMap[token]
	if !found {
		return false
	}
	if perm.mode != O_WRONLY {
		return false
	}
	if !(perm.namespace == namespace && perm.repository == repository) {
		return false
	}
	return true
}
