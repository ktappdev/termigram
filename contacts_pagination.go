package main

import (
	"fmt"
	"strings"
)

const defaultContactsPageSize = 20

type contactPage struct {
	Items      []ContactOutput
	Offset     int
	Limit      int
	Total      int
	Page       int
	TotalPages int
}

func paginateContacts(contacts []ContactOutput, offset int, limit int) contactPage {
	if limit <= 0 {
		limit = defaultContactsPageSize
	}
	if offset < 0 {
		offset = 0
	}

	totalPages := 0
	page := 0
	if len(contacts) > 0 {
		totalPages = (len(contacts) + limit - 1) / limit
		maxOffset := (totalPages - 1) * limit
		if offset > maxOffset {
			offset = maxOffset
		}
		page = (offset / limit) + 1
	}

	end := offset + limit
	if end > len(contacts) {
		end = len(contacts)
	}

	items := make([]ContactOutput, 0, end-offset)
	if offset < end {
		items = append(items, contacts[offset:end]...)
	}

	return contactPage{
		Items:      items,
		Offset:     offset,
		Limit:      limit,
		Total:      len(contacts),
		Page:       page,
		TotalPages: totalPages,
	}
}

func contactDisplayName(c ContactOutput) string {
	name := strings.TrimSpace(c.FirstName + " " + c.LastName)
	if name != "" {
		return name
	}
	if c.Username != "" {
		return "@" + c.Username
	}
	return fmt.Sprintf("User %d", c.UserID)
}

func contactTarget(c ContactOutput) string {
	if c.Username != "" {
		return "@" + c.Username
	}
	return fmt.Sprintf("%d", c.UserID)
}

func cachedChatFromContact(c ContactOutput) CachedChat {
	return CachedChat{
		Label:  contactDisplayName(c),
		Target: contactTarget(c),
	}
}
