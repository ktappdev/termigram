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
	cli.mu.Lock()
	defer cli.mu.Unlock()
	if at.IsZero() {
		at = time.Now()
	}
	cli.chatLastActivity[normalized] = at
	if strings.TrimSpace(message) != "" {
		cli.chatLastMessage[normalized] = strings.TrimSpace(message)
	}
}

func (cli *TelegramCLI) setChatUnreadCount(target string, unreadCount int) {
	normalized := normalizeUsername(target)
	if normalized == "" {
		return
	}
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.chatUnreadCount[normalized] = unreadCount
}

func (cli *TelegramCLI) clearChatUnreadCount(target string) {
	normalized := normalizeUsername(target)
	if normalized == "" {
		return
	}
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.chatUnreadCount[normalized] = 0
}

func (cli *TelegramCLI) incrementChatUnreadCount(target string) int {
	normalized := normalizeUsername(target)
	if normalized == "" {
		return 0
	}
	cli.mu.Lock()
	defer cli.mu.Unlock()
	cli.chatUnreadCount[normalized]++
	return cli.chatUnreadCount[normalized]
}

func (cli *TelegramCLI) listCachedChats(limit int) []CachedChat {
	cli.mu.RLock()
	defer cli.mu.RUnlock()

	chats := make([]CachedChat, 0, len(cli.users))
	seen := make(map[string]struct{}, len(cli.users))
	for _, u := range cli.users {
		label := strings.TrimSpace(u.FirstName + " " + u.LastName)
		if label == "" {
			label = u.Username
		}
		if label == "" {
			label = fmt.Sprintf("User %d", u.ID)
		}

		target := fmt.Sprintf("%d", u.ID)
		if u.Username != "" {
			target = "@" + u.Username
		}

		normalized := normalizeUsername(target)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		chats = append(chats, CachedChat{
			Label:        label,
			Target:       target,
			LastMessage:  cli.chatLastMessage[normalized],
			LastActivity: cli.chatLastActivity[normalized],
			UnreadCount:  cli.chatUnreadCount[normalized],
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
