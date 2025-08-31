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
	"github.com/whynot00/imsi_bot/internal/telegram/middleware/matches"
)

type Bot struct {
	b       *bot.Bot
	handler *handler.Handler
	mw      *middleware.Middleware
	fsm     *fsm.FSM
}

func New(ctx context.Context, db *sqlx.DB, token string, obsRepo observation.Repository, whlRepo whitelist.Repository) *Bot {

	machine := fsm.New(ctx)
	middleware := middleware.Create(whlRepo)

	b, _ := bot.New(token,
		bot.WithMiddlewares(fsm.Middleware(machine), middleware.Whitelist),
	)

	return &Bot{
		b:       b,
		handler: handler.Create(db, obsRepo, whlRepo),
		mw:      middleware,
		fsm:     machine,
	}

}

func (b *Bot) InitRoutes() *Bot {

	regFetch, _ := regexp.Compile(`^\d{14,15}$`)
	b.b.RegisterHandlerRegexp(bot.HandlerTypeMessageText, regFetch, b.handler.FetchObservation, fsm.WithStates(fsm.StateDefault))

	// обновление БД
	b.b.RegisterHandler(bot.HandlerTypeMessageText, "update", bot.MatchTypeCommand, b.handler.DownloadUpdateCommand, b.mw.Whitelist, fsm.WithStates(fsm.StateDefault))
	b.b.RegisterHandlerMatchFunc(matches.MatchUpdate, b.handler.DownloadUpdate, fsm.WithStates(states.UploadFileState))

	// создание пользователя
	b.b.RegisterHandler(bot.HandlerTypeMessageText, "add_user", bot.MatchTypeCommand, b.handler.CreateUserCommand, fsm.WithStates(fsm.StateDefault))
	b.b.RegisterHandler(bot.HandlerTypeMessageText, "", bot.MatchTypeContains, b.handler.CreateUser, fsm.WithStates(states.CreateUserState))

	// кнопки
	//
	// отмена
	b.b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "cancel_button", bot.MatchTypeContains, b.handler.CancelButton, fsm.WithStates(fsm.StateAny))
	return b
}

func (b *Bot) Start(ctx context.Context) {

	b.b.Start(ctx)
}
