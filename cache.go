package main

import (
	"strings"

	"github.com/gotd/td/tg"
)

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimPrefix(username, "@"))
}

func (cli *TelegramCLI) cacheUser(user *tg.User) {
	cli.mu.Lock()
	defer cli.mu.Unlock()

	if oldUsername, ok := cli.usernameByUserID[user.ID]; ok && oldUsername != "" {
		delete(cli.usersByName, oldUsername)
	}

	cli.users[user.ID] = user

	username := normalizeUsername(user.Username)
	if username == "" {
		delete(cli.usernameByUserID, user.ID)
		return
	}

	cli.usersByName[username] = user
	cli.usernameByUserID[user.ID] = username
}

func (cli *TelegramCLI) getUserByID(userID int64) (*tg.User, bool) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	user, found := cli.users[userID]
	return user, found
}

func (cli *TelegramCLI) getUserByUsername(username string) (*tg.User, bool) {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	user, found := cli.usersByName[normalizeUsername(username)]
	return user, found
}
