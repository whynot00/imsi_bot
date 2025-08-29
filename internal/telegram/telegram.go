package telegram

import (
	"context"
	"fmt"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jmoiron/sqlx"
	"github.com/whynot00/imsi_bot/internal/domain/observation"
	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
	"github.com/whynot00/imsi_bot/internal/telegram/handler"
	"github.com/whynot00/imsi_bot/internal/telegram/middleware"
)

type Bot struct {
	b       *bot.Bot
	handler *handler.Handler
	mw      *middleware.Middleware
}

func New(ctx context.Context, db *sqlx.DB, token string, obsRepo observation.Repository, whlRepo whitelist.Repository) *Bot {

	b, _ := bot.New(token,
		bot.WithDefaultHandler(
			func(ctx context.Context, bot *bot.Bot, update *models.Update) {
				fmt.Printf("%s: %d\n", update.Message.From.Username, update.Message.From.ID)
			},
		),
	)

	return &Bot{
		b:       b,
		handler: handler.Create(db, obsRepo, whlRepo),
		mw:      middleware.Create(whlRepo),
	}

}

func (b *Bot) InitRoutes() *Bot {

	reg, err := regexp.Compile(`^\d{14,15}$`)
	if err != nil {
		panic(err)
	}
	b.b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, reg, b.handler.FetchObservation, b.mw.Whitelist)

	b.b.RegisterHandler(bot.HandlerTypeMessageText, "new", bot.MatchTypeCommand, b.handler.CreateUser, b.mw.Whitelist)

	b.b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeContains, b.handler.DownloadUpdate)
	return b
}

func (b *Bot) Start(ctx context.Context) {

	b.b.Start(ctx)
}
