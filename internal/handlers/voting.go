package handlers

import (
	"context"
	"fmt"
	"log"
	"questions-vote/internal/models"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// sendVoteQuestions sends a pair of questions for voting
func (h *BotHandler) sendVoteQuestions(chatID int64) error {
	questions, err := h.questionService.GetQuestions()
	if err != nil {
		return fmt.Errorf("failed to get questions: %w", err)
	}

	if len(questions) < 2 {
		return fmt.Errorf("not enough questions available")
	}

	q1, q2 := questions[0], questions[1]

	err = h.sendQuestion(chatID, q1, 1)
	if err != nil {
		return fmt.Errorf("failed to send first question: %w", err)
	}

	err = h.sendQuestion(chatID, q2, 2)
	if err != nil {
		return fmt.Errorf("failed to send second question: %w", err)
	}

	// Send voting keyboard
	keyboard := h.createVoteKeyboard(q1.ID, q2.ID)
	h.rateLimiter.Record(chatID)

	_, err = h.bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:      tu.ID(chatID),
		Text:        "Какой вопрос лучше?",
		ReplyMarkup: keyboard,
	})

	return err
}

// sendQuestion sends a single formatted question
func (h *BotHandler) sendQuestion(chatID int64, question *models.Question, number int) error {
	if question.HandoutImg != "" {
		_, err := h.bot.SendPhoto(context.Background(), &telego.SendPhotoParams{
			ChatID: tu.ID(chatID),
			Photo:  tu.FileFromBytes([]byte(question.HandoutImg), "handout.jpg"),
		})
		if err != nil {
			log.Printf("Failed to send handout image: %v", err)
		}
	}

	questionText := h.formatQuestion(question, number)
	_, err := h.bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:    tu.ID(chatID),
		Text:      questionText,
		ParseMode: telego.ModeHTML,
		LinkPreviewOptions: &telego.LinkPreviewOptions{
			IsDisabled: true,
		},
	})

	return err
}

// createVoteKeyboard creates inline keyboard for voting
func (h *BotHandler) createVoteKeyboard(q1ID, q2ID int) *telego.InlineKeyboardMarkup {
	return &telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				{
					Text:         "Первый",
					CallbackData: fmt.Sprintf("vote_%d_%d_1", q1ID, q2ID),
				},
				{
					Text:         "Второй",
					CallbackData: fmt.Sprintf("vote_%d_%d_2", q1ID, q2ID),
				},
			},
			{
				{
					Text:         "Не могу выбрать",
					CallbackData: fmt.Sprintf("vote_%d_%d_0", q1ID, q2ID),
				},
			},
		},
	}
}

// handleCallback handles button callback queries
func (h *BotHandler) handleCallback(bot *telego.Bot, update telego.Update) {
	query := update.CallbackQuery

	// Parse callback data: "vote_q1id_q2id_choice"
	parts := strings.Split(query.Data, "_")
	if len(parts) != 4 {
		log.Printf("Invalid callback data: %s", query.Data)
		return
	}

	q1ID, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Printf("Invalid q1ID: %s", parts[1])
		return
	}

	q2ID, err := strconv.Atoi(parts[2])
	if err != nil {
		log.Printf("Invalid q2ID: %s", parts[2])
		return
	}

	choice, err := strconv.Atoi(parts[3])
	if err != nil {
		log.Printf("Invalid choice: %s", parts[3])
		return
	}

	var selectedID *int
	switch choice {
	case 1:
		selectedID = &q1ID
	case 2:
		selectedID = &q2ID
	}
	// choice == 0 means skip (selectedID remains nil)

	err = h.voteService.SaveVote(query.From.ID, q1ID, q2ID, selectedID)
	if err != nil {
		log.Printf("Failed to save vote: %v", err)
	}

	response := h.getConfirmationMessage(q1ID, q2ID, selectedID)
	if selectedID != nil {
		statsMessage := h.getQuestionStatsMessage(q1ID, q2ID, selectedID)
		response += " " + statsMessage
	}

	err = bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
	})
	if err != nil {
		log.Printf("Failed to answer callback query: %v", err)
	}

	_, err = bot.EditMessageText(context.Background(), &telego.EditMessageTextParams{
		ChatID:    tu.ID(query.Message.GetChat().ID),
		MessageID: query.Message.GetMessageID(),
		Text:      response,
	})
	if err != nil {
		log.Printf("Failed to edit message: %v", err)
	}

	chatID := query.Message.GetChat().ID
	err = h.sendVoteQuestions(chatID)
	if err != nil {
		log.Printf("Failed to send next vote questions: %v", err)
		_, sendErr := bot.SendMessage(context.Background(), &telego.SendMessageParams{
			ChatID: tu.ID(chatID),
			Text:   "Произошла ошибка при получении следующих вопросов. Попробуйте команду /vote снова.",
		})
		if sendErr != nil {
			log.Printf("Failed to send error message: %v", sendErr)
		}
	}
}
