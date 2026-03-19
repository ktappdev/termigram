package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

func (cli *TelegramCLI) fetchContacts(ctx context.Context) ([]ContactOutput, error) {
	contacts, err := cli.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return nil, err
	}

	out := make([]ContactOutput, 0)
	switch c := contacts.(type) {
	case *tg.ContactsContacts:
		out = make([]ContactOutput, 0, len(c.Users))
		for _, user := range c.Users {
			u, ok := user.(*tg.User)
			if !ok {
				continue
			}
			cli.cacheUser(u)
			out = append(out, ContactOutput{
				UserID:    u.ID,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Username:  u.Username,
				Phone:     u.Phone,
			})
		}
	default:
		return nil, fmt.Errorf("unexpected contacts response type: %T", contacts)
	}

	return out, nil
}

func (cli *TelegramCLI) loadContacts(ctx context.Context) error {
	contacts, err := cli.fetchContacts(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s %d contacts\n", green("✓ Loaded"), len(contacts))
	return nil
}

func (cli *TelegramCLI) showSelf(ctx context.Context) {
	self, err := cli.client.Self(ctx)
	if err != nil {
		fmt.Printf("%s %v\n", red("Error getting user info:"), err)
		return
	}

	fmt.Println()
	fmt.Println(bold(cyan("Your Account")))
	fmt.Printf("  %s %d\n", dim("ID:"), self.ID)
	fmt.Printf("  %s %s\n", dim("Name:"), strings.TrimSpace(self.FirstName+" "+self.LastName))
	fmt.Printf("  %s @%s\n", dim("Username:"), self.Username)
	fmt.Printf("  %s %s\n", dim("Phone:"), self.Phone)
}
