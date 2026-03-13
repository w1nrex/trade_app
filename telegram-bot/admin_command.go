package telegrambot

import (
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func AdminCommand(b *gotgbot.Bot, ctx *ext.Context) error {
	id := ctx.EffectiveUser.Id
	if id != 1234567890 {
		_, err := ctx.EffectiveMessage.Reply(b, "You arent an admin. U cant use this command.", &gotgbot.SendMessageOpts{
			ParseMode: "html",
		})
		if err != nil {
			log.Printf("Didnt work command AdminCommand: %s", err)
		}
	}

	text := `"Hello, Alexander.\n This is statistic for the day."`

	_, err := ctx.EffectiveMessage.Reply(b, text, &gotgbot.SendMessageOpts{
		ParseMode: "html",
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{{
				{Text: "Statistic", CallbackData: "Statistic"},
			}},
		},
	})
	if err != nil {
		log.Printf("Didnt work ctx.EffectivfeMessage.Reply in AdminCommand with button: %s", err)
	}

	return err
}
