package formatter

import (
	"fmt"
	"html"
	"time"

	"github.com/whynot00/imsi_bot/internal/domain/errorx"
	"github.com/whynot00/imsi_bot/internal/domain/observation"
)

func ObservationMessage(o *observation.Observation) string {

	str := fmt.Sprintf(
		"📡 *Наблюдение*\n"+
			"`Стандарт : %-8s`\n"+
			"`Сигнал   : %d db`\n"+
			"`Дата     : %s`\n"+
			"`Коорд.   : %.6f, %.6f`\n",
		o.Standart,
		o.SignalStrength,
		o.Date.UTC().Format("2006-01-02 15:04:05"),
		o.LAT, o.LON,
	)

	if o.Operator != "" {
		str = fmt.Sprintf("%s`Оператор : %-12s`\n", str, o.Operator)
	}

	if o.IMSI != "" {
		str = fmt.Sprintf("%s`IMSI     : %s`\n", str, o.IMSI)
	}

	if o.IMEI != "" {
		str = fmt.Sprintf("%s`IMEI     : %s`\n", str, o.IMEI)
	}

	return str
}

func NoMatchesMessage() string {
	return "\nСовпадений нет\n"
}

func UserAlreadyExistsByID(id int64) string {
	return fmt.Sprintf("⚠️ Пользователь с ID *%d* уже существует", id)
}

func InvalidID(id string) string {
	return fmt.Sprintf("❌ Неверный ID: *%s*\\.\nПроверьте ввод и попробуйте снова\\.", id)
}

func UserAddedByID(id string) string {
	return fmt.Sprintf("✅ Пользователь с ID *%s* успешно добавлен", id)
}

func InternalError() string {
	return "❌ Внутренняя сервера. Попробуйте позже."
}

func InternalErrorAdmin(e errorx.ReqError) string {
	ts := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")

	req := e.Request
	if len(req) > 2000 { // на всякий случай, чтобы не поломать сообщение
		req = req[:2000] + "…"
	}

	errText := ""
	if e.Err != nil {
		errText = e.Err.Error()
	}

	return fmt.Sprintf(
		"⚠️ <b>Internal Server Error</b>\n"+
			"🕒 %s\n"+
			"👤 UserID: <code>%d</code>\n"+
			"📨 Запрос:\n<pre>%s</pre>\n"+
			"💥 Ошибка:\n<pre>%s</pre>",
		ts,
		e.UserID,
		html.EscapeString(req),
		html.EscapeString(errText),
	)
}
