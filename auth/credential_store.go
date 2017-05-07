// Ref: https://github.com/rqlite/rqlite/blob/master/auth/credential_store.go
// Package auth is a lightweight credential store.
// It provides functionality for loading credentials, as well as validating credentials.
package auth

import (
	"encoding/json"
	"io"
)

// Credential represents authentication and authorization configuration for a single user.
type Credential struct {
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Perms    []string `json:"perms,omitempty"`
}

// CredentialsStore stores authentication and authorization information for all users.
type CredentialsStore struct {
	store map[string]string
	perms map[string]map[string]bool
}

// NewCredentialsStore returns a new instance of a CredentialStore.
func NewCredentialsStore() *CredentialsStore {
	return &CredentialsStore{
		store: make(map[string]string),
		perms: make(map[string]map[string]bool),
	}
}

// Load loads credential information from a reader.
func (c *CredentialsStore) Load(r io.Reader) error {
	dec := json.NewDecoder(r)
	// Read open bracket
	_, err := dec.Token()
	if err != nil {
		return err
	}

	var cred Credential
	for dec.More() {
		err := dec.Decode(&cred)
		if err != nil {
			return err
		}
		c.store[cred.Username] = cred.Password
		c.perms[cred.Username] = make(map[string]bool, len(cred.Perms))
		for _, p := range cred.Perms {
			c.perms[cred.Username][p] = true
		}
	}

	// Read closing bracket.
	_, err = dec.Token()
	if err != nil {
		return err
	}

	return nil
}

// Check returns true if the password is correct for the given username.
func (c *CredentialsStore) Check(username, password string) bool {
	pw, ok := c.store[username]
	return ok && password == pw
}

// HasPerm returns true if username has the given perm. It does not
// perform any password checking.
func (c *CredentialsStore) HasPerm(username string, perm string) bool {
	m, ok := c.perms[username]
	if !ok {
		return false
	}
	if _, ok := m[perm]; !ok {
		return false
	}
	return true
}
