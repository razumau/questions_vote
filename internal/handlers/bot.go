package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"questions-vote/internal/services"
	"questions-vote/pkg/ratelimiter"
	"time"

	"github.com/mymmrac/telego"
)

// BotHandler handles all bot operations
type BotHandler struct {
	bot             *telego.Bot
	questionService *services.QuestionService
	voteService     *services.VoteService
	rateLimiter     *ratelimiter.RateLimiter
}

// NewBotHandler creates a new bot handler
func NewBotHandler() (*BotHandler, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN environment variable is required")
	}

	bot, err := telego.NewBot(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &BotHandler{
		bot:             bot,
		questionService: services.NewQuestionService(),
		voteService:     services.NewVoteService(),
		rateLimiter:     ratelimiter.New(5 * time.Second), // 5 second cooldown
	}, nil
}

// GetQuestionsCount returns the total number of questions
func (h *BotHandler) GetQuestionsCount() (int, error) {
	count, err := h.questionService.GetQuestionsCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get questions count: %w", err)
	}
	return count, nil
}

// Run starts the bot
func (h *BotHandler) Run() error {
	updates, err := h.bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to get updates: %w", err)
	}

	log.Println("Bot is running...")

	for update := range updates {
		go h.processUpdate(update)
	}

	return nil
}

// processUpdate processes a single update
func (h *BotHandler) processUpdate(update telego.Update) {
	if update.Message != nil {
		switch update.Message.Text {
		case "/start":
			h.handleStart(h.bot, update)
		case "/vote":
			h.handleVote(h.bot, update)
		}
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Data != "" &&
			len(update.CallbackQuery.Data) > 5 &&
			update.CallbackQuery.Data[:5] == "vote_" {
			h.handleCallback(h.bot, update)
		}
	}
}
