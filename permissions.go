package main

import (
	"encoding/json"
	"errors"
	"strings"
)

const (
	permissionsDS = "core/permission.ds"
)

type Permission struct {
	Owner   string `json:"owner"`
	Public  string `json:"public,omitempty"`
	Friend  string `json:"friend,omitempty"`
	Private string `json:"private,omitempty"`
}

func permissions(resource string) (*Permission, error) {
	//Docs are open to public
	if isDocPath(resource) {
		return &Permission{
			Owner:  "",
			Public: "r",
		}, nil

	}
	ds := openCoreDS(permissionsDS)
	key, err := json.Marshal(resource)
	if err != nil {
		logError(err)
		return nil, err
	}

	var value []byte
	var perm *Permission

	err = ds.Get(key, value)
	if err != nil {
		logError(err)
		return nil, err
	}
	if value == nil {
		return nil, errors.New("Permissions not found for resource: " + resource)
	}

	err = json.Unmarshal(value, perm)
	if err != nil {
		logError(err)
		return nil, err
	}

	return perm, nil
}

func setPermissions(resource string, permissions *Permission) error {
	err := permissions.validate()
	if err != nil {
		return err
	}
	ds := openCoreDS(permissionsDS)
	key, err := json.Marshal(resource)
	if err != nil {
		logError(err)
		return err
	}

	value, err := json.Marshal(permissions)
	if err != nil {
		logError(err)
		return err
	}

	err = ds.Put(key, value)
	if err != nil {
		logError(err)
		return err
	}

	return nil
}

func (p *Permission) validate() error {
	strings.ToLower(p.Private)
	strings.ToLower(p.Friend)
	strings.ToLower(p.Public)
	if !validPermission(p.Private) {
		return publicError(errors.New("Invalid Private permission. Value must be blank, r, w, or rw"))
	}
	if !validPermission(p.Friend) {
		return publicError(errors.New("Invalid Friend permission. Value must be blank, r, w, or rw"))
	}

	if !validPermission(p.Public) {
		return publicError(errors.New("Invalid Public permission. Value must be blank, r, w, or rw"))
	}
	return nil
}

func validPermission(permission string) bool {
	switch permission {
	case "", "r", "w", "rw":
		return true
	default:
		return false
	}

}

func (p *Permission) canPermission(permType rune, u *User) bool {
	if u == nil { //public
		if strings.ContainsRune(p.Public, permType) {
			return true
		}
		return false
	}
	if p.isOwner(u) {
		if strings.ContainsRune(p.Private, permType) {
			return true
		}
		return false
	}

	//authenticated friend
	if strings.ContainsRune(p.Friend, permType) {
		return true
	}
	return false

}

func (p *Permission) canRead(u *User) bool {
	return p.canPermission('r', u)
}

func (p *Permission) canWrite(u *User) bool {
	return p.canPermission('w', u)
}

func (p *Permission) isOwner(u *User) bool {
	if p.Owner == u.username {
		return true
	}

	return false
}

func isDocPath(resource string) bool {
	root, _ := splitRootAndPath(resource)
	return root == "docs"
}
