package main

import "context"

// TelegramBackend defines the shared operations used by one-shot CLI commands.
type TelegramBackend interface {
	IsAuthorized(ctx context.Context) (bool, error)
	GetSelf(ctx context.Context) (*UserOutput, error)
	SendMessage(ctx context.Context, target string, text string) error
}

// HistoryBackend is implemented by backends that can fetch recent messages.
type HistoryBackend interface {
	GetMessages(ctx context.Context, target string, limit int) ([]MessageOutput, error)
}

// ContactsBackend is implemented by backends that can list contacts.
type ContactsBackend interface {
	GetContacts(ctx context.Context) ([]ContactOutput, error)
}

// UsernameDiscoveryBackend is implemented by backends that can resolve users
// and support username prefix lookup.
type UsernameDiscoveryBackend interface {
	FindUserByUsername(ctx context.Context, username string) (*UserOutput, error)
	FindMatchingUsernames(prefix string, limit int) []string
}

// TargetResolverBackend resolves a CLI target (user ID or @username) to user
// metadata for richer output.
type TargetResolverBackend interface {
	ResolveTarget(ctx context.Context, target string) (*UserOutput, error)
}

// UserCacheBackend is implemented by backends that maintain user caches.
type UserCacheBackend interface {
	CacheUser(user *UserOutput)
}

// Output formats for CLI commands.
type CLIOutput struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UserOutput represents user data for JSON output.
type UserOutput struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

// MessageOutput represents message data for JSON output.
type MessageOutput struct {
	ID       int64  `json:"id"`
	FromID   int64  `json:"from_id"`
	FromName string `json:"from_name"`
	Message  string `json:"message"`
	Date     int64  `json:"date"`
}

// ContactOutput represents contact data for JSON output.
type ContactOutput struct {
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username,omitempty"`
	Phone     string `json:"phone,omitempty"`
}
