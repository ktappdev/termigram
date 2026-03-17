package main

import (
	"context"
	"fmt"

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
		fmt.Printf("Loaded %d contacts\n", len(c.Users))
	}

	return nil
}

func (cli *TelegramCLI) showContacts(ctx context.Context) {
	fmt.Println("\n--- Contacts ---")

	contacts, err := cli.api.ContactsGetContacts(ctx, 0)
	if err != nil {
		fmt.Printf("Error getting contacts: %v\n", err)
		return
	}

	switch c := contacts.(type) {
	case *tg.ContactsContacts:
		for _, user := range c.Users {
			if u, ok := user.(*tg.User); ok {
				cli.cacheUser(u)
				username := ""
				if u.Username != "" {
					username = fmt.Sprintf(" (@%s)", u.Username)
				}
				fmt.Printf("%d: %s %s%s\n", u.ID, u.FirstName, u.LastName, username)
			}
		}
		if len(c.Users) == 0 {
			fmt.Println("No contacts found")
		}
	default:
		fmt.Printf("Unexpected response type: %T\n", contacts)
	}

	fmt.Println("----------------")
}

func (cli *TelegramCLI) showSelf(ctx context.Context) {
	self, err := cli.client.Self(ctx)
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		return
	}

	fmt.Println("\n--- Your Account ---")
	fmt.Printf("ID: %d\n", self.ID)
	fmt.Printf("First Name: %s\n", self.FirstName)
	fmt.Printf("Last Name: %s\n", self.LastName)
	fmt.Printf("Username: @%s\n", self.Username)
	fmt.Printf("Phone: %s\n", self.Phone)
	fmt.Println("--------------------")
}
