package handlers

import (
	"fmt"
	"log"
	"os"
	"questions-vote/internal/models"
	"questions-vote/internal/services"
	"questions-vote/pkg/ratelimiter"
	"strconv"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
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

// SetupHandlers configures all bot handlers
func (h *BotHandler) SetupHandlers() {
	bh, _ := th.NewBotHandler(h.bot, nil)

	bh.Handle(h.handleStart, th.CommandEqual("start"))
	bh.Handle(h.handleVote, th.CommandEqual("vote"))
	bh.Handle(h.handleCallback, th.CallbackDataPrefix("vote_"))

	go bh.Start()
}

// GetQuestionsCount returns the total number of questions
func (h *BotHandler) GetQuestionsCount() int {
	// TODO: Get from tournament service
	return 1000 // Mock value
}

// Run starts the bot
func (h *BotHandler) Run() error {
	h.SetupHandlers()
	
	log.Println("Bot is running...")
	
	// Keep the bot running
	select {}
}