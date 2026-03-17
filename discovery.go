package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
)

// findUserByIDOrUsername looks up a user by ID or @username
func (cli *TelegramCLI) findUserByIDOrUsername(ctx context.Context, idOrUsername string) (*tg.User, error) {
	if strings.HasPrefix(idOrUsername, "@") {
		username := normalizeUsername(idOrUsername)

		if user, found := cli.getUserByUsername(username); found {
			return user, nil
		}

		resolved, err := cli.api.ContactsResolveUsername(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("user @%s not found", username)
		}

		for _, user := range resolved.Users {
			if u, ok := user.(*tg.User); ok {
				cli.cacheUser(u)
				return u, nil
			}
		}

		return nil, fmt.Errorf("user @%s not found", username)
	}

	userID, err := strconv.ParseInt(idOrUsername, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID or username: %s", idOrUsername)
	}

	if user, found := cli.getUserByID(userID); found {
		return user, nil
	}

	return &tg.User{ID: userID}, nil
}

// findMatchingUsernames searches usersByName cache for usernames starting with prefix
func (cli *TelegramCLI) findMatchingUsernames(prefix string, limit int) []string {
	prefix = normalizeUsername(prefix)
	if prefix == "" {
		return nil
	}

	cli.mu.RLock()
	defer cli.mu.RUnlock()

	var matches []string
	for username := range cli.usersByName {
		if strings.HasPrefix(username, prefix) {
			matches = append(matches, "@"+username)
		}
	}
	sort.Strings(matches)
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}
