package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// handleStart handles the /start command
func (h *BotHandler) handleStart(bot *telego.Bot, update telego.Update) {
	questionsCount, err := h.GetQuestionsCount()
	if err != nil {
		log.Printf("Failed to get questions count: %v", err)
		// Send error message to user
		_, sendErr := bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: tu.ID(update.Message.Chat.ID),
			Text:   "Извините, произошла ошибка при получении информации о турнире. Попробуйте позже.",
		})
		if sendErr != nil {
			log.Printf("Failed to send error message: %v", sendErr)
		}
		return
	}

	message := fmt.Sprintf(
		"Вспомнить хорошие вопросы спустя год сложно, поэтому я предлагаю вам сравнить все "+
			"вопросы за 2022 год (%d штук). Бот будет присылать пары вопросов, и вам нужно будет выбрать, "+
			"какой из двух лучше. Цель этого этапа — выбрать шортлист из 100 вопросов. Вопросы с низким (Эло-подобным) "+
			"рейтингом будут постепенно убираться, но не раньше, чем каждый поучаствует в пяти матчах.",
		questionsCount,
	)

	_, err = bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		Text:   message,
	})

	if err != nil {
		log.Printf("Failed to send start message: %v", err)
	}
}

// handleVote handles the /vote command
func (h *BotHandler) handleVote(bot *telego.Bot, update telego.Update) {
	chatID := update.Message.Chat.ID

	// Check rate limiting
	canSendIn := h.rateLimiter.CanSendInSeconds(chatID)
	if canSendIn > 0 {
		log.Printf("Rate limited for chat %d, can send in %d seconds", chatID, canSendIn)
		_, err := bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: tu.ID(chatID),
			Text:   fmt.Sprintf("Пожалуйста, подождите %d секунд перед следующим голосованием.", canSendIn),
		})
		if err != nil {
			log.Printf("Failed to send rate limit message: %v", err)
		}
		return
	}

	err := h.sendVoteQuestions(chatID)
	if err != nil {
		log.Printf("Failed to send vote questions: %v", err)
		_, sendErr := bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: tu.ID(chatID),
			Text:   "Извините, произошла ошибка при получении вопросов. Попробуйте позже.",
		})
		if sendErr != nil {
			log.Printf("Failed to send error message: %v", sendErr)
		}
	}
}
