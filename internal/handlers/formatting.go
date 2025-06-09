package handlers

import (
	"fmt"
	"questions-vote/internal/models"
	"strings"
)

// formatQuestion formats a question for display
func (h *BotHandler) formatQuestion(question *models.Question, number int) string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("<b>Вопрос %d.</b>", number))
	
	// Add handout if exists
	if question.HandoutStr != "" {
		parts = append(parts, fmt.Sprintf("<b>Раздаточный материал</b>:\n%s", question.HandoutStr))
	}
	
	parts = append(parts, question.Question)
	
	// Add spoiler section with answer details
	spoilerParts := []string{
		fmt.Sprintf("<b>Ответ</b>: %s", question.Answer),
	}
	
	if question.AcceptedAnswer != "" {
		spoilerParts = append(spoilerParts, fmt.Sprintf("<b>Зачёт</b>: %s", question.AcceptedAnswer))
	}
	
	spoilerParts = append(spoilerParts, fmt.Sprintf("<b>Комментарий</b>: %s", question.Comment))
	spoilerParts = append(spoilerParts, fmt.Sprintf("<b>Источник</b>: %s", question.Source))
	
	spoilerContent := strings.Join(spoilerParts, "\n")
	parts = append(parts, fmt.Sprintf("<tg-spoiler>\n%s\n</tg-spoiler>", spoilerContent))
	
	return strings.Join(parts, "\n")
}

// getConfirmationMessage returns confirmation message after voting
func (h *BotHandler) getConfirmationMessage(q1ID, q2ID int, selectedID *int) string {
	if selectedID == nil {
		return "Ок, сейчас пришлём другую пару вопросов."
	}
	
	if *selectedID == q1ID {
		return "Записали, что первый вопрос лучше."
	} else if *selectedID == q2ID {
		return "Записали, что второй вопрос лучше."
	}
	
	return "Голос записан."
}

// getQuestionStatsMessage returns statistics message for questions
func (h *BotHandler) getQuestionStatsMessage(q1ID, q2ID int, selectedID *int) string {
	if selectedID == nil {
		return ""
	}
	
	stats, err := h.voteService.GetQuestionStats(q1ID, q2ID)
	if err != nil {
		return ""
	}
	
	if len(stats) < 2 {
		return ""
	}
	
	first := stats[0]
	second := stats[1]
	
	firstPct := 0.0
	if first.Matches > 0 {
		firstPct = float64(first.Wins) / float64(first.Matches) * 100
	}
	
	secondPct := 0.0
	if second.Matches > 0 {
		secondPct = float64(second.Wins) / float64(second.Matches) * 100
	}
	
	return fmt.Sprintf(
		"У первого теперь %s в %s (%.1f%%), а у второго — %s в %s (%.1f%%).",
		h.inflectWins(first.Wins),
		h.inflectMatches(first.Matches),
		firstPct,
		h.inflectWins(second.Wins),
		h.inflectMatches(second.Matches),
		secondPct,
	)
}

// inflectWins returns properly inflected word for wins in Russian
func (h *BotHandler) inflectWins(number int) string {
	var winsWord string
	
	if number%100 >= 11 && number%100 <= 19 {
		winsWord = "побед"
	} else if number%10 == 1 {
		winsWord = "победа"
	} else if number%10 >= 2 && number%10 <= 4 {
		winsWord = "победы"
	} else {
		winsWord = "побед"
	}
	
	return fmt.Sprintf("%d %s", number, winsWord)
}

// inflectMatches returns properly inflected word for matches in Russian
func (h *BotHandler) inflectMatches(number int) string {
	var matchesWord string
	
	if number == 1 {
		matchesWord = "матче"
	} else if number == 0 {
		matchesWord = "матчей"
	} else {
		matchesWord = "матчах"
	}
	
	return fmt.Sprintf("%d %s", number, matchesWord)
}