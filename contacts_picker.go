package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func (cli *TelegramCLI) printContactPage(page contactPage) {
	fmt.Println()
	fmt.Println(bold(cyan("Contacts")))
	if page.Total == 0 {
		fmt.Println(dim("No contacts found."))
		return
	}
	fmt.Printf("%s %d-%d of %d (page %d/%d)\n", dim("Showing"), page.Offset+1, page.Offset+len(page.Items), page.Total, page.Page, page.TotalPages)
	for _, contact := range page.Items {
		fmt.Printf("  %s %s %s\n", blue("•"), bold(contactDisplayName(contact)), dim("("+contactTarget(contact)+")"))
	}
}

func (cli *TelegramCLI) showContacts(ctx context.Context) {
	contacts, err := cli.fetchContacts(ctx)
	if err != nil {
		fmt.Printf("%s %v\n", red("Error getting contacts:"), err)
		return
	}

	if len(contacts) == 0 {
		fmt.Println()
		fmt.Println(bold(cyan("Contacts")))
		fmt.Println(dim("No contacts found."))
		return
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		cli.printContactPage(paginateContacts(contacts, 0, defaultContactsPageSize))
		return
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Printf("%s %v\n", red("Could not start contacts browser:"), err)
		cli.printContactPage(paginateContacts(contacts, 0, defaultContactsPageSize))
		return
	}
	defer term.Restore(fd, oldState)

	out := func(format string, args ...any) {
		fmt.Printf(strings.ReplaceAll(format, "\n", "\r\n"), args...)
	}

	pageSize := defaultContactsPageSize
	offset := 0
	selected := 0

	render := func() contactPage {
		page := paginateContacts(contacts, offset, pageSize)
		if selected >= len(page.Items) {
			selected = len(page.Items) - 1
		}
		if selected < 0 {
			selected = 0
		}

		fmt.Print("\033[2J\033[H")
		out("%s %s\n", bold(cyan("Contacts")), dim("(↑/↓ select, n/p page, Enter switch, Esc cancel)"))
		if page.Total == 0 {
			out("%s\n", dim("No contacts found."))
			return page
		}
		out("%s %d-%d of %d (page %d/%d)\n\n", dim("Showing"), page.Offset+1, page.Offset+len(page.Items), page.Total, page.Page, page.TotalPages)
		for i, contact := range page.Items {
			cursor := "  "
			lineStyle := func(s string) string { return s }
			if i == selected {
				cursor = yellow("➤ ")
				lineStyle = bold
			}
			out("%s%s %s\n", cursor, lineStyle(contactDisplayName(contact)), dim("("+contactTarget(contact)+")"))
		}
		return page
	}

	page := render()
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			break
		}

		switch b := buf[0]; b {
		case 3:
			fmt.Print("\033[2J\033[H")
			out("\n")
			return
		case 13:
			if len(page.Items) == 0 {
				continue
			}
			fmt.Print("\033[2J\033[H")
			out("\n")
			cli.activateCachedChat(cachedChatFromContact(page.Items[selected]), false)
			return
		case 'n', 'N':
			if offset+pageSize < len(contacts) {
				offset += pageSize
				selected = 0
			}
		case 'p', 'P':
			if offset-pageSize >= 0 {
				offset -= pageSize
				selected = 0
			}
		case 27:
			next, ok := readByteWithTimeout(fd, 15)
			if !ok {
				fmt.Print("\033[2J\033[H")
				out("\n")
				fmt.Println(dim("Contacts cancelled."))
				return
			}
			if next != '[' {
				fmt.Print("\033[2J\033[H")
				out("\n")
				fmt.Println(dim("Contacts cancelled."))
				return
			}
			arrow, ok := readByteWithTimeout(fd, 15)
			if !ok {
				fmt.Print("\033[2J\033[H")
				out("\n")
				fmt.Println(dim("Contacts cancelled."))
				return
			}
			switch arrow {
			case 'A':
				if selected > 0 {
					selected--
				}
			case 'B':
				if selected < len(page.Items)-1 {
					selected++
				}
			case 'C':
				if offset+pageSize < len(contacts) {
					offset += pageSize
					selected = 0
				}
			case 'D':
				if offset-pageSize >= 0 {
					offset -= pageSize
					selected = 0
				}
			default:
				fmt.Print("\033[2J\033[H")
				out("\n")
				fmt.Println(dim("Contacts cancelled."))
				return
			}
		}
		page = render()
	}

	fmt.Print("\033[2J\033[H")
	out("\n")
}
