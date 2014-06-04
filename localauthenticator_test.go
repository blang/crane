package main

import (
	"testing"
)

func TestAuthenticate(t *testing.T) {
	const USER, PASS = "user", "pass"
	auth := &LocalAuthenticator{}
	if !auth.Authenticate(USER, PASS) {
		t.Errorf("Should authenticate every user, denied for %s:%s", USER, PASS)
	}
}

func TestAuthorize(t *testing.T) {
	auth := NewLocalAuthenticator()
	token, granted := auth.Authorize("user", "pass", "user", "repository", []string{"123", "456"}, O_WRONLY)
	if !granted {
		t.Fatal("Did not authorize user on its own namespace")
	}
	if !(len(token) > 5) {
		t.Errorf("Authorization token was to short: %s", token)
	}

	token, granted = auth.Authorize("user", "pass", "anotheruser", "repository", []string{"123", "456"}, O_WRONLY)
	if granted {
		t.Fatal("Did authorize user on another than its own namespace")
	}
}

func TestPerms(t *testing.T) {
	auth := NewLocalAuthenticator()
	var (
		USER       = "user"
		PASS       = "pass"
		NAMESPACE  = USER
		REPOSITORY = "REPOSITORY"
		IMG1       = "123"
		IMG2       = "456"
		IMAGES     = []string{IMG1, IMG2}
	)
	writeToken, _ := auth.Authorize(USER, PASS, NAMESPACE, REPOSITORY, IMAGES, O_WRONLY)
	readToken, _ := auth.Authorize(USER, PASS, NAMESPACE, REPOSITORY, IMAGES, O_RDONLY)

	// Push image

	// Valid image
	if !auth.HasPermPushImage(writeToken, "123") {
		t.Error("Write token does not grant push image")
	}
	if auth.HasPermPushImage(readToken, "123") {
		t.Error("Read token grants push image")
	}

	// Invalid image
	if auth.HasPermPushImage(writeToken, "invalid") {
		t.Error("Write token grants push to invalid image")
	}

	// Pull image

	// Valid image
	if auth.HasPermPullImage(writeToken, "123") {
		t.Error("Write token grants pull image")
	}
	if !auth.HasPermPullImage(readToken, "123") {
		t.Error("Read token does not grant pull image")
	}

	// Invalid image
	if auth.HasPermPullImage(readToken, "invalid") {
		t.Error("Read token grants pull of invalid image")
	}

	// Push tag

	if !auth.HasPermPushTag(writeToken, NAMESPACE, REPOSITORY, IMG1, "testtag") {
		t.Error("Write token does not grant push tag")
	}
	if auth.HasPermPushTag(readToken, NAMESPACE, REPOSITORY, IMG1, "testtag") {
		t.Error("Read token grants push tag")
	}
	// Invalid image
	if auth.HasPermPushTag(writeToken, NAMESPACE, REPOSITORY, "invalidimage", "testtag") {
		t.Error("Write token grants push tag of invalid image")
	}
	// Invalid repository
	if auth.HasPermPushTag(writeToken, NAMESPACE, "anotherrepo", IMG1, "testtag") {
		t.Error("Write token grants push tag of invalid repository")
	}
	// Invalid namespace
	if auth.HasPermPushTag(writeToken, "anothernamespace", REPOSITORY, IMG1, "testtag") {
		t.Error("Write token grants push tag of invalid namespace")
	}

	// Pull tag

	if auth.HasPermPullTag(writeToken, NAMESPACE, REPOSITORY, "testtag") {
		t.Error("Write token grants pull tag")
	}
	if !auth.HasPermPullTag(readToken, NAMESPACE, REPOSITORY, "testtag") {
		t.Error("Read token does not grant pull tag")
	}
	// Invalid repository
	if auth.HasPermPullTag(readToken, NAMESPACE, "anotherrepo", "testtag") {
		t.Error("Read token grants pull tag of invalid repository")
	}
	// Invalid namespace
	if auth.HasPermPullTag(readToken, "anothernamespace", REPOSITORY, "testtag") {
		t.Error("Read token grants pull tag of invalid namespace")
	}

	// Pull tags

	if auth.HasPermPullTags(writeToken, NAMESPACE, REPOSITORY) {
		t.Error("Write token grants pull tags")
	}
	if !auth.HasPermPullTags(readToken, NAMESPACE, REPOSITORY) {
		t.Error("Read token does not grant pull tags")
	}
	// Invalid repository
	if auth.HasPermPullTags(readToken, NAMESPACE, "anotherrepo") {
		t.Error("Read token grants pull tags of invalid repository")
	}
	// Invalid namespace
	if auth.HasPermPullTags(readToken, "anothernamespace", REPOSITORY) {
		t.Error("Read token grants pull tags of invalid namespace")
	}

	// Push checksums
	if !auth.HasPermPushChecksums(writeToken, NAMESPACE, REPOSITORY) {
		t.Error("Write token does not grant push checksums")
	}
	if auth.HasPermPushChecksums(readToken, NAMESPACE, REPOSITORY) {
		t.Error("Read token grants push checksums")
	}
	// Invalid repository
	if auth.HasPermPushChecksums(writeToken, NAMESPACE, "anotherrepo") {
		t.Error("Read token grants push checksums to invalid repository")
	}
	// Invalid namespace
	if auth.HasPermPushChecksums(writeToken, "anothernamespace", REPOSITORY) {
		t.Error("Read token grants push checksums to invalid namespace")
	}
}
