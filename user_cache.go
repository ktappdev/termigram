package main

import (
	"sync"

	"github.com/gotd/td/tg"
)

// UserCache stores Telegram user records and provides username/ID lookups.
type UserCache struct {
	mu sync.RWMutex

	users            map[int64]*tg.User
	usersByName      map[string]*tg.User
	usernameByUserID map[int64]string
}

func newUserCache() *UserCache {
	return &UserCache{
		users:            make(map[int64]*tg.User),
		usersByName:      make(map[string]*tg.User),
		usernameByUserID: make(map[int64]string),
	}
}
