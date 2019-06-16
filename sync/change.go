package sync

import "strings"

type Change uint8

const (
	ChangeLocalAdd Change = 1 << iota
	ChangeLocalDelete
	ChangeLocalUpdate
	ChangeRemoteAdd
	ChangeRemoteDelete
	ChangeRemoteUpdate
)

var changeNameMap = map[Change]string{
	ChangeLocalAdd:     "local-add",
	ChangeLocalDelete:  "local-delete",
	ChangeLocalUpdate:  "local-update",
	ChangeRemoteAdd:    "remote-add",
	ChangeRemoteDelete: "remote-delete",
	ChangeRemoteUpdate: "remote-update",
}

func (c Change) Changed() bool {
	return c != 0
}

func (c *Change) Register(add Change) {
	*c = *c | add
}

func (c Change) HasAll(test ...Change) bool {
	for _, t := range test {
		if c&t == 0 {
			return false
		}
	}

	return true
}

func (c Change) HasOne(test ...Change) bool {
	for _, t := range test {
		if c&t != 0 {
			return true
		}
	}
	return false
}

func (c Change) Is(test Change) bool {
	return c == test
}

func (c Change) String() string {
	if !c.Changed() {
		return "none"
	}

	names := []string{}
	for k, v := range changeNameMap {
		if c.HasOne(k) {
			names = append(names, v)
		}
	}

	return strings.Join(names, ", ")
}
