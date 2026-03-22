package main

import "testing"

func TestImagePickerItemsUsesNewestImagesFirst(t *testing.T) {
	items := imagePickerItems([]legacyTranscriptEntry{
		{
			MessageID: 1,
			Header:    "Alice",
			Meta:      "10:00:00",
			Body:      "[image #1] older.png\nfirst caption",
			Image: &ImageAttachment{
				Kind:     imageKind,
				Name:     "older.png",
				MIMEType: "image/png",
			},
		},
		{
			MessageID: 2,
			Header:    "You",
			Meta:      "10:01:00  ✓",
			Body:      "[image #2] newer.jpg\nsecond caption",
			Image: &ImageAttachment{
				Kind:     imageKind,
				Name:     "newer.jpg",
				MIMEType: "image/jpeg",
			},
		},
	}, 10)

	if len(items) != 2 {
		t.Fatalf("expected 2 picker items, got %d", len(items))
	}
	if items[0].Title != "newer.jpg" {
		t.Fatalf("expected newest image first, got %q", items[0].Title)
	}
	if items[1].Title != "older.png" {
		t.Fatalf("expected older image second, got %q", items[1].Title)
	}
	if items[0].Subtitle == "" {
		t.Fatalf("expected subtitle for newest image")
	}
}

func TestFilterImagePickerItemsMatchesCaptionAndSender(t *testing.T) {
	items := []imagePickerItem{
		{Title: "meme.png", Subtitle: "Alice • 10:00:00 • funny cat", QueryText: "meme.png Alice funny cat"},
		{Title: "receipt.jpg", Subtitle: "You • 10:01:00", QueryText: "receipt.jpg You"},
	}

	if got := filterImagePickerItems(items, "cat"); len(got) != 1 || got[0].Title != "meme.png" {
		t.Fatalf("expected caption query to match meme image, got %#v", got)
	}

	if got := filterImagePickerItems(items, "alice"); len(got) != 1 || got[0].Title != "meme.png" {
		t.Fatalf("expected sender query to match meme image, got %#v", got)
	}
}
