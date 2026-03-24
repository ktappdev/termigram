package main

import "strings"

func legacyChatHeaderRows(label string, target string, width int, pendingReply string) []string {
	rows := []string{
		dim(truncateVisibleWidth("Active chat: "+strings.TrimSpace(label)+" ("+strings.TrimSpace(target)+")", width)),
		dim(strings.Repeat("─", maxInt(width, 1))),
	}
	if banner := strings.TrimSpace(pendingReply); banner != "" {
		rows = append(rows, dim(truncateVisibleWidth("↪ "+banner, width)))
	}
	return rows
}
