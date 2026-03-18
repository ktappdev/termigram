package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CLICommand represents a non-interactive CLI command.
type CLICommand struct {
	Name    string
	Args    []string
	JSON    bool
	Limit   int
	Timeout time.Duration
}

// RunCLICommand executes a non-interactive CLI command against a backend.
func RunCLICommand(ctx context.Context, backend TelegramBackend, cmd CLICommand) error {
	switch cmd.Name {
	case "send":
		return cmdSend(ctx, backend, cmd.Args, cmd.JSON)
	case "get":
		return cmdGet(ctx, backend, cmd.Args, cmd.Limit, cmd.JSON)
	case "contacts":
		return cmdContacts(ctx, backend, cmd.JSON)
	case "me":
		return cmdMe(ctx, backend, cmd.JSON)
	case "find":
		return cmdFind(ctx, backend, cmd.Args, cmd.JSON)
	default:
		return fmt.Errorf("unknown command: %s", cmd.Name)
	}
}

func cmdSend(ctx context.Context, backend TelegramBackend, args []string, asJSON bool) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: send <user_id|@username> <message>")
	}

	target := args[0]
	text := strings.Join(args[1:], " ")
	resolvedUser, _ := resolveTargetForOutput(ctx, backend, target)

	if err := backend.SendMessage(ctx, target, text); err != nil {
		return err
	}

	if asJSON {
		data := map[string]interface{}{
			"target":    target,
			"message":   text,
			"timestamp": time.Now().Unix(),
		}
		if resolvedUser != nil {
			sentTo := resolvedUser.Username
			if sentTo == "" {
				sentTo = strings.TrimSpace(resolvedUser.FirstName + " " + resolvedUser.LastName)
			}
			data["sent_to"] = sentTo
			data["user_id"] = resolvedUser.ID
		}
		output := CLIOutput{Success: true, Data: data}
		jsonOut, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonOut))
		return nil
	}

	displayTarget := target
	if resolvedUser != nil {
		displayTarget = resolvedUser.Username
		if displayTarget == "" {
			displayTarget = strings.TrimSpace(resolvedUser.FirstName + " " + resolvedUser.LastName)
		}
		if displayTarget == "" {
			displayTarget = target
		}
	}
	fmt.Printf("Message sent to %s!\n", displayTarget)
	return nil
}

func cmdGet(ctx context.Context, backend TelegramBackend, args []string, limit int, asJSON bool) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get <user_id|@username> [--limit N]")
	}

	if limit <= 0 {
		limit = 10
	}

	historyBackend, ok := backend.(HistoryBackend)
	if !ok {
		return fmt.Errorf("backend does not support message history retrieval")
	}

	target := args[0]
	resolvedUser, _ := resolveTargetForOutput(ctx, backend, target)
	messages, err := historyBackend.GetMessages(ctx, target, limit)
	if err != nil {
		return err
	}
	messages = chronologicalMessages(messages)

	if asJSON {
		data := map[string]interface{}{
			"target":   target,
			"count":    len(messages),
			"messages": messages,
		}
		if resolvedUser != nil {
			data["user"] = resolvedUser.Username
			data["user_id"] = resolvedUser.ID
		}
		output := CLIOutput{Success: true, Data: data}
		jsonOut, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonOut))
		return nil
	}

	displayTarget := target
	if resolvedUser != nil {
		displayTarget = resolvedUser.Username
		if displayTarget == "" {
			displayTarget = strings.TrimSpace(resolvedUser.FirstName + " " + resolvedUser.LastName)
		}
		if displayTarget == "" {
			displayTarget = target
		}
	}

	fmt.Printf("Last %d messages from %s:\n", len(messages), displayTarget)
	fmt.Println(strings.Repeat("-", 50))
	for _, msg := range messages {
		dateStr := time.Unix(msg.Date, 0).Format("2006-01-02 15:04:05")
		fmt.Printf("[%s] %s: %s\n", dateStr, msg.FromName, msg.Message)
	}
	return nil
}

func cmdContacts(ctx context.Context, backend TelegramBackend, asJSON bool) error {
	contactsBackend, ok := backend.(ContactsBackend)
	if !ok {
		return fmt.Errorf("backend does not support contacts")
	}

	contacts, err := contactsBackend.GetContacts(ctx)
	if err != nil {
		return err
	}

	if asJSON {
		output := CLIOutput{Success: true, Data: map[string]interface{}{
			"count":    len(contacts),
			"contacts": contacts,
		}}
		jsonOut, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonOut))
		return nil
	}

	fmt.Println("\n--- Contacts ---")
	for _, c := range contacts {
		username := ""
		if c.Username != "" {
			username = fmt.Sprintf(" (@%s)", c.Username)
		}
		fmt.Printf("%d: %s %s%s\n", c.UserID, c.FirstName, c.LastName, username)
	}
	if len(contacts) == 0 {
		fmt.Println("No contacts found")
	}
	fmt.Println("----------------")
	return nil
}

func cmdMe(ctx context.Context, backend TelegramBackend, asJSON bool) error {
	self, err := backend.GetSelf(ctx)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	if asJSON {
		output := CLIOutput{Success: true, Data: self}
		jsonOut, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonOut))
		return nil
	}

	fmt.Println("\n--- Your Account ---")
	fmt.Printf("ID: %d\n", self.ID)
	fmt.Printf("First Name: %s\n", self.FirstName)
	fmt.Printf("Last Name: %s\n", self.LastName)
	fmt.Printf("Username: @%s\n", self.Username)
	fmt.Printf("Phone: %s\n", self.Phone)
	fmt.Println("--------------------")
	return nil
}

func cmdFind(ctx context.Context, backend TelegramBackend, args []string, asJSON bool) error {
	_ = ctx
	if len(args) < 1 {
		return fmt.Errorf("usage: find <prefix>")
	}

	discoveryBackend, ok := backend.(UsernameDiscoveryBackend)
	if !ok {
		return fmt.Errorf("backend does not support username prefix matching")
	}

	prefix := args[0]
	matches := discoveryBackend.FindMatchingUsernames(prefix, 10)

	if asJSON {
		output := CLIOutput{Success: true, Data: map[string]interface{}{
			"prefix":  prefix,
			"count":   len(matches),
			"matches": matches,
		}}
		jsonOut, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonOut))
		return nil
	}

	if len(matches) == 0 {
		fmt.Printf("No cached usernames found starting with %q.\n", prefix)
	} else {
		fmt.Println("Cached matches:")
		for _, m := range matches {
			fmt.Println(" ", m)
		}
	}
	return nil
}

func resolveTargetForOutput(ctx context.Context, backend TelegramBackend, target string) (*UserOutput, error) {
	resolver, ok := backend.(TargetResolverBackend)
	if !ok {
		return nil, nil
	}
	return resolver.ResolveTarget(ctx, target)
}

func chronologicalMessages(messages []MessageOutput) []MessageOutput {
	ordered := append([]MessageOutput(nil), messages...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Date == ordered[j].Date {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].Date < ordered[j].Date
	})
	return ordered
}

// parseLimitArg parses --limit argument from args.
func parseLimitArg(args []string) ([]string, int) {
	for i, arg := range args {
		if arg == "--limit" && i+1 < len(args) {
			if limit, err := strconv.Atoi(args[i+1]); err == nil && limit > 0 {
				newArgs := make([]string, 0, len(args)-2)
				newArgs = append(newArgs, args[:i]...)
				newArgs = append(newArgs, args[i+2:]...)
				return newArgs, limit
			}
		}
	}
	return args, 10
}
