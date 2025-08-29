package telegram

import (
	"context"
	"regexp"

	"github.com/go-telegram/bot"
	"github.com/jmoiron/sqlx"
	fsm "github.com/whynot00/go-telegram-fsm"
	"github.com/whynot00/imsi_bot/internal/domain/observation"
	"github.com/whynot00/imsi_bot/internal/domain/states"
	"github.com/whynot00/imsi_bot/internal/domain/whitelist"
	"github.com/whynot00/imsi_bot/internal/telegram/handler"
	"github.com/whynot00/imsi_bot/internal/telegram/middleware"
)

type Bot struct {
	b       *bot.Bot
	handler *handler.Handler
	mw      *middleware.Middleware
	fsm     *fsm.FSM
}

func New(ctx context.Context, db *sqlx.DB, token string, obsRepo observation.Repository, whlRepo whitelist.Repository) *Bot {

	machine := fsm.New(ctx)

	b, _ := bot.New(token,
		bot.WithMiddlewares(fsm.Middleware(machine)),
	)

	return &Bot{
		b:       b,
		handler: handler.Create(db, obsRepo, whlRepo),
		mw:      middleware.Create(whlRepo),
		fsm:     machine,
	}

}

func (b *Bot) InitRoutes() *Bot {

	reg, _ := regexp.Compile(`^\d{14,15}$`)
	b.b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, reg, b.handler.FetchObservation, b.mw.Whitelist, fsm.WithStates(fsm.StateDefault))

	// создание пользователя
	b.b.RegisterHandler(bot.HandlerTypeMessageText, "add_user", bot.MatchTypeCommand, b.handler.CreateUserCommand, b.mw.Whitelist, fsm.WithStates(fsm.StateDefault))
	b.b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeContains, b.handler.CreateUser, b.mw.Whitelist, fsm.WithStates(states.CreateUserState))

	b.b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeContains, b.handler.DownloadUpdate)
	return b
}

func (b *Bot) Start(ctx context.Context) {

	b.b.Start(ctx)
}
