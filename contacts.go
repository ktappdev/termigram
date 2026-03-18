package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

func (cli *TelegramCLI) loadContacts(ctx context.Context) error {
	contacts, err := cli.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		return err
	}

	switch c := contacts.(type) {
	case *tg.ContactsContacts:
		for _, user := range c.Users {
			if u, ok := user.(*tg.User); ok {
				cli.cacheUser(u)
			}
		}
		fmt.Printf("%s %d contacts\n", green("✓ Loaded"), len(c.Users))
	}

	return nil
}

func (cli *TelegramCLI) showContacts(ctx context.Context) {
	fmt.Println()
	fmt.Println(bold(cyan("Contacts")))

	contacts, err := cli.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		fmt.Printf("%s %v\n", red("Error getting contacts:"), err)
		return
	}

	switch c := contacts.(type) {
	case *tg.ContactsContacts:
		for _, user := range c.Users {
			if u, ok := user.(*tg.User); ok {
				cli.cacheUser(u)
				name := strings.TrimSpace(u.FirstName + " " + u.LastName)
				if name == "" {
					name = fmt.Sprintf("User %d", u.ID)
				}
				handle := fmt.Sprintf("id:%d", u.ID)
				if u.Username != "" {
					handle = "@" + u.Username
				}
				fmt.Printf("  %s %s %s\n", blue("•"), bold(name), dim("("+handle+")"))
			}
		}
		if len(c.Users) == 0 {
			fmt.Println(dim("No contacts found"))
		}
	default:
		fmt.Printf("%s %T\n", yellow("Unexpected response type:"), contacts)
	}
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
