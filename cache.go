package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gotd/td/tg"
)

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimPrefix(username, "@"))
}

func buildChatLabel(user *tg.User) (label string, target string) {
	label = strings.TrimSpace(user.FirstName + " " + user.LastName)
	if label == "" {
		label = user.Username
	}
	if label == "" {
		label = fmt.Sprintf("User %d", user.ID)
	}

	target = fmt.Sprintf("%d", user.ID)
	if user.Username != "" {
		target = "@" + user.Username
	}

	return label, target
}

func (cli *TelegramCLI) cacheUser(user *tg.User) {
	cli.userCache.mu.Lock()
	defer cli.userCache.mu.Unlock()

	if oldUsername, ok := cli.userCache.usernameByUserID[user.ID]; ok && oldUsername != "" {
		delete(cli.userCache.usersByName, oldUsername)
	}

	cli.userCache.users[user.ID] = user

	username := normalizeUsername(user.Username)
	if username == "" {
		delete(cli.userCache.usernameByUserID, user.ID)
		return
	}

	cli.userCache.usersByName[username] = user
	cli.userCache.usernameByUserID[user.ID] = username
}

func (cli *TelegramCLI) getUserByID(userID int64) (*tg.User, bool) {
	cli.userCache.mu.RLock()
	defer cli.userCache.mu.RUnlock()

	user, found := cli.userCache.users[userID]
	return user, found
}

func (cli *TelegramCLI) getUserByUsername(username string) (*tg.User, bool) {
	cli.userCache.mu.RLock()
	defer cli.userCache.mu.RUnlock()

	user, found := cli.userCache.usersByName[normalizeUsername(username)]
	return user, found
}

type CachedChat struct {
	Label        string
	Target       string
	LastMessage  string
	LastActivity time.Time
	UnreadCount  int
}

func (cli *TelegramCLI) markChatActivity(target string, message string, at time.Time) {
	normalized := normalizeUsername(target)
	if normalized == "" {
		return
	}
	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	if at.IsZero() {
		at = time.Now()
	}
	cli.chatState.chatLastActivity[normalized] = at
	if strings.TrimSpace(message) != "" {
		cli.chatState.chatLastMessage[normalized] = strings.TrimSpace(message)
	}
}

func (cli *TelegramCLI) setChatUnreadCount(target string, unreadCount int) {
	normalized := normalizeUsername(target)
	if normalized == "" {
		return
	}
	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	cli.chatState.chatUnreadCount[normalized] = unreadCount
}

func (cli *TelegramCLI) clearChatUnreadCount(target string) {
	normalized := normalizeUsername(target)
	if normalized == "" {
		return
	}
	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	cli.chatState.chatUnreadCount[normalized] = 0
}

func (cli *TelegramCLI) incrementChatUnreadCount(target string) int {
	normalized := normalizeUsername(target)
	if normalized == "" {
		return 0
	}
	cli.chatState.mu.Lock()
	defer cli.chatState.mu.Unlock()
	cli.chatState.chatUnreadCount[normalized]++
	return cli.chatState.chatUnreadCount[normalized]
}

func (cli *TelegramCLI) listCachedChats(limit int) []CachedChat {
	cli.chatState.mu.RLock()
	cli.userCache.mu.RLock()
	defer cli.userCache.mu.RUnlock()
	defer cli.chatState.mu.RUnlock()

	chats := make([]CachedChat, 0, len(cli.userCache.users))
	seen := make(map[string]struct{}, len(cli.userCache.users))
	for _, u := range cli.userCache.users {
		label, target := buildChatLabel(u)

		normalized := normalizeUsername(target)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		chats = append(chats, CachedChat{
			Label:        label,
			Target:       target,
			LastMessage:  cli.chatState.chatLastMessage[normalized],
			LastActivity: cli.chatState.chatLastActivity[normalized],
			UnreadCount:  cli.chatState.chatUnreadCount[normalized],
		})
	}

	sort.Slice(chats, func(i, j int) bool {
		if chats[i].LastActivity.Equal(chats[j].LastActivity) {
			return strings.ToLower(chats[i].Label) < strings.ToLower(chats[j].Label)
		}
		return chats[i].LastActivity.After(chats[j].LastActivity)
	})
	if limit > 0 && len(chats) > limit {
		chats = chats[:limit]
	}
	return chats
}
