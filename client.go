package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

type TelegramCLI struct {
	client             *telegram.Client
	api                *tg.Client
	sender             *message.Sender
	ctx                context.Context
	cancel             context.CancelFunc
	reader             *bufio.Reader
	nowFunc          func() time.Time
	userCache        *UserCache
	chatState        *ChatState
	httpClient       *http.Client
	transcriptStore  *TranscriptStore
	envLookup        func(string) string
	ttyCheck                func() bool
	interactiveResumeDialogRefresher func(ctx context.Context, cli *TelegramCLI, limit int) ([]CachedChat, error)
	sendPreparedImageWithBackend    func(ctx context.Context, backend *UserBackend, target string, prepared preparedImageSource, caption string, opts SendOptions) (int64, error)
	fetchReplyMessagesFunc          func(ctx context.Context, cli *TelegramCLI, ids []int64) (map[int64]*tg.Message, error)
	imageDownloadFunc               func(ctx context.Context, cli *TelegramCLI, entry transcriptEntry, path string) error
	openLocalPath                   func(path string) error
	cacheOutboundImageFunc          func(target string, attachment *ImageAttachment, sourcePath string) (string, error)
	config                          Config
}

func NewTelegramCLI(appID int, appHash string, sessionPath string) *TelegramCLI {
	return NewTelegramCLIWithConfig(Config{
		TelegramAppID:   appID,
		TelegramAppHash: appHash,
		SessionPath:     sessionPath,
	})
}

func NewTelegramCLIWithConfig(cfg Config) *TelegramCLI {
	appID := cfg.TelegramAppID
	appHash := cfg.TelegramAppHash
	sessionPath := cfg.SessionPath
	sessionDir := filepath.Dir(sessionPath)
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		fmt.Printf("Warning: could not create session directory: %v\n", err)
	}

	cli := &TelegramCLI{
		reader:                          bufio.NewReader(os.Stdin),
		nowFunc:                         time.Now,
		httpClient:                      &http.Client{Timeout: 30 * time.Second},
		envLookup:                       os.Getenv,
		ttyCheck:                        interactiveTTYAvailable,
		interactiveResumeDialogRefresher: func(ctx context.Context, cli *TelegramCLI, limit int) ([]CachedChat, error) {
			return cli.fetchDialogs(ctx, limit, false)
		},
		sendPreparedImageWithBackend: func(ctx context.Context, backend *UserBackend, target string, prepared preparedImageSource, caption string, opts SendOptions) (int64, error) {
			return backend.sendPreparedImage(ctx, target, prepared, caption, opts)
		},
		fetchReplyMessagesFunc:           fetchReplyMessages,
		imageDownloadFunc:                downloadImageAttachment,
		openLocalPath:                    defaultOpenLocalPath,
		cacheOutboundImageFunc:           cacheOutboundImageCopy,
		userCache:                       newUserCache(),
		chatState:                       newChatState(),
		config:                          cfg,
	}
	cli.transcriptStore = newTranscriptStore(&cli.config)

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

func (cli *TelegramCLI) promptInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := cli.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (cli *TelegramCLI) RunInteractive() error {
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

func (cli *TelegramCLI) handleUpdates(_ context.Context, updates tg.UpdatesClass) error {

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
		cli.writeOutput("[info] Too many updates received; waiting for fresh updates...")
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
	cli.printMessage(messageFromShortUserUpdate(update))
}

func (cli *TelegramCLI) Run() error {
	return cli.RunInteractive()
}

func (cli *TelegramCLI) processShortChatMessage(update *tg.UpdateShortChatMessage) {
	if update.GetOut() {
		return
	}
	cli.printMessage(messageFromShortChatUpdate(update))
}
