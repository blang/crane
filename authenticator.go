package main

type Mode int

const (
	O_RDONLY Mode = iota
	O_WRONLY
)

type Authenticator interface {
	Authenticate(user string, pass string) bool
	Authorize(user, pass, namespace, repository string, imageIDs []string, mode Mode) (string, bool)
	HasPermPushImage(token string, imageID string) bool
	HasPermPullImage(token string, imageID string) bool
	HasPermPushTag(token string, namespace, repository string, imageID string, tag string) bool
	HasPermPullTag(token string, namespace, repository string, tag string) bool
	HasPermPullTags(token string, namespace, repository string) bool
	HasPermPushChecksums(token string, namespace, repository string) bool
}
