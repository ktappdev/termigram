package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

type TelegramCLI struct {
	client            *telegram.Client
	api               *tg.Client
	sender            *message.Sender
	ctx               context.Context
	cancel            context.CancelFunc
	reader            *bufio.Reader
	mu                sync.RWMutex
	users             map[int64]*tg.User
	usersByName       map[string]*tg.User // username -> user mapping
	usernameByUserID  map[int64]string
	currentChatTarget string
	currentChatLabel  string
	chatLastActivity  map[string]time.Time
	chatLastMessage   map[string]string
	chatUnreadCount   map[string]int
	seenIncoming      map[string]time.Time
	legacyMu          sync.RWMutex
	legacyConsole     *legacyConsole
	legacyTranscripts map[string][]legacyTranscriptEntry
	legacyLoaded      map[string]bool
}

func NewTelegramCLI(appID int, appHash string, sessionPath string) *TelegramCLI {
	sessionDir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		fmt.Printf("Warning: could not create session directory: %v\n", err)
	}

	cli := &TelegramCLI{
		reader:            bufio.NewReader(os.Stdin),
		users:             make(map[int64]*tg.User),
		usersByName:       make(map[string]*tg.User),
		usernameByUserID:  make(map[int64]string),
		chatLastActivity:  make(map[string]time.Time),
		chatLastMessage:   make(map[string]string),
		chatUnreadCount:   make(map[string]int),
		seenIncoming:      make(map[string]time.Time),
		legacyTranscripts: make(map[string][]legacyTranscriptEntry),
		legacyLoaded:      make(map[string]bool),
	}

	cli.client = telegram.NewClient(appID, appHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: sessionPath,
		},
		UpdateHandler: telegram.UpdateHandlerFunc(func(ctx context.Context, updates tg.UpdatesClass) error {
			return cli.handleUpdates(ctx, updates)
		}),
	})

	return cli
}

func (cli *TelegramCLI) RunLegacy() error {
	ctx, cancel := context.WithCancel(context.Background())
	cli.ctx = ctx
	cli.cancel = cancel

	fmt.Println("Connecting to Telegram...")

	return cli.client.Run(ctx, func(ctx context.Context) error {
		cli.api = cli.client.API()
		cli.sender = message.NewSender(cli.client.API())

		status, err := cli.client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}

		if !status.Authorized {
			fmt.Println("Not authenticated. Starting auth flow...")
			authFlow := auth.NewFlow(
				&UserAuthenticator{cli: cli},
				auth.SendCodeOptions{},
			)
			if err := cli.client.Auth().IfNecessary(ctx, authFlow); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}
		}

		self, err := cli.client.Self(ctx)
		if err != nil {
			return fmt.Errorf("failed to get self: %w", err)
		}
		cli.cacheUser(self)

		fmt.Printf("\n%s %s\n", green("● Connected"), bold("Telegram MTProto"))
		fmt.Printf("%s %s %s\n", dim("Logged in as"), bold(strings.TrimSpace(self.FirstName+" "+self.LastName)), dim("@"+self.Username))
		fmt.Printf("%s %d\n\n", dim("User ID:"), self.ID)

		printInteractiveStartup()
		printHelp()

		if err := cli.loadContacts(ctx); err != nil {
			fmt.Printf("Warning: could not load contacts: %v\n", err)
		}
		if _, err := cli.fetchDialogs(ctx, 50, false); err != nil {
			fmt.Printf("Warning: could not load dialogs: %v\n", err)
		}

		cli.commandLoop(ctx)
		return nil
	})
}

func (cli *TelegramCLI) handleUpdates(ctx context.Context, updates tg.UpdatesClass) error {
	_ = ctx

	switch u := updates.(type) {
	case *tg.Updates:
		cli.cacheUsersFromClasses(u.GetUsers())
		cli.processUpdateList(u.GetUpdates())
	case *tg.UpdatesCombined:
		cli.cacheUsersFromClasses(u.GetUsers())
		cli.processUpdateList(u.GetUpdates())
	case *tg.UpdateShort:
		cli.processSingleUpdate(u.GetUpdate())
	case *tg.UpdateShortMessage:
		cli.processShortUserMessage(u)
	case *tg.UpdateShortChatMessage:
		cli.processShortChatMessage(u)
	case *tg.UpdatesTooLong:
		cli.writeLegacyOutput("[info] Too many updates received; waiting for fresh updates...")
	}

	return nil
}

func (cli *TelegramCLI) cacheUsersFromClasses(users []tg.UserClass) {
	for _, user := range users {
		if usr, ok := user.(*tg.User); ok {
			cli.cacheUser(usr)
		}
	}
}

func (cli *TelegramCLI) processUpdateList(updates []tg.UpdateClass) {
	for _, update := range updates {
		cli.processSingleUpdate(update)
	}
}

func (cli *TelegramCLI) processSingleUpdate(update tg.UpdateClass) {
	switch upd := update.(type) {
	case *tg.UpdateNewMessage:
		if msg, ok := upd.GetMessage().(*tg.Message); ok {
			if msg.GetOut() {
				return
			}
			cli.printMessage(msg)
		}
	case *tg.UpdateNewChannelMessage:
		if msg, ok := upd.GetMessage().(*tg.Message); ok {
			if msg.GetOut() {
				return
			}
			cli.printMessage(msg)
		}
	}
}

func (cli *TelegramCLI) processShortUserMessage(update *tg.UpdateShortMessage) {
	if update.GetOut() {
		return
	}

	msg := &tg.Message{
		ID:      update.GetID(),
		Date:    update.GetDate(),
		Message: update.GetMessage(),
		FromID: &tg.PeerUser{
			UserID: update.GetUserID(),
		},
	}
	cli.printMessage(msg)
}

func (cli *TelegramCLI) Run() error {
	return cli.RunLegacy()
}

func (cli *TelegramCLI) processShortChatMessage(update *tg.UpdateShortChatMessage) {
	if update.GetOut() {
		return
	}

	msg := &tg.Message{
		ID:      update.GetID(),
		Date:    update.GetDate(),
		Message: update.GetMessage(),
		FromID: &tg.PeerUser{
			UserID: update.GetFromID(),
		},
	}
	cli.printMessage(msg)
}
