package main

import "fmt"

func printHelp() {
	fmt.Println()
	fmt.Println(bold(cyan("Commands")))
	fmt.Println("  \\me                 Show current user info")
	fmt.Println("  \\contacts           Browse contacts by page and switch on selection")
	fmt.Println("  \\find <query>       Find cached chats/usernames and switch via selector")
	fmt.Println("  \\msg <id|@user> <text>  Send message and enter chat mode")
	fmt.Println("  \\image <source> [caption]  Send an image into the active chat")
	fmt.Println("  \\openimage [last|message-id|query]  Open/pick an image from the active chat")
	fmt.Println("  \\to <id|@user>      Switch active chat")
	fmt.Println("  \\here               Show active chat")
	fmt.Println("  \\chats              Interactive recent chats picker (↑/↓, Enter, Esc, filter)")
	fmt.Println("  \\unread             Pick from chats with unread messages and enter the selection")
	fmt.Println("  \\close              Exit chat mode")
	fmt.Println("  \\chat / \\back      Deprecated aliases for \\here/\\to and \\close")
	fmt.Println("  \\help               Show this help")
	fmt.Println("  \\quit               Exit")
	fmt.Println()
}

func printInteractiveStartup() {
	fmt.Println(dim("Quick start"))
	fmt.Println("  \\to @user           switch chats")
	fmt.Println("  \\image ./meme.png   send an image into the active chat")
	fmt.Println("  \\openimage          pick from recent images in the active chat")
	fmt.Println("  \\openimage last     open the newest image fast")
	fmt.Println("  inline preview      shows visible chat images inline in kitty/iTerm2 when supported")
	fmt.Println("  \\help               show the full command list")
	fmt.Println()
}
