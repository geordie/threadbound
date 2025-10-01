package texgen

import (
	"os"
	"strings"
	"testing"
	"time"

	"threadbound/internal/models"
)

// getMockData returns a small set of mock messages, handles, and reactions for testing
func getMockData() ([]models.Message, map[int]models.Handle, map[string][]models.Reaction) {
	// Mock handles (contacts)
	handles := map[int]models.Handle{
		1: {
			ID:          1,
			Service:     "iMessage",
			Contact:     "+14039090086",
			DisplayName: "+14039090086",
		},
		2: {
			ID:          2,
			Service:     "iMessage",
			Contact:     "+16047904258",
			DisplayName: "+16047904258",
		},
	}

	// Mock messages
	baseTime := time.Date(2025, 9, 27, 10, 0, 0, 0, time.UTC)

	text1 := "Hey, how's the fishing going?"
	text2 := "Great! Caught a nice salmon today"
	text3 := "That's awesome! 🎣"
	text4 := "Thanks! Want to join me tomorrow?"
	text5 := "Absolutely! What time?"

	// Helper function to get handle ID pointer
	handleID1 := 1
	handleID2 := 2

	messages := []models.Message{
		{
			ID:             1,
			GUID:           "MSG-001-GUID-RECEIVED",
			Text:           &text1,
			Date:           int64(baseTime.UnixNano()),
			FormattedDate:  baseTime,
			IsFromMe:       false,
			HandleID:       &handleID1,
			HasAttachments: false,
		},
		{
			ID:             2,
			GUID:           "MSG-002-GUID-SENT",
			Text:           &text2,
			Date:           int64(baseTime.Add(5 * time.Minute).UnixNano()),
			FormattedDate:  baseTime.Add(5 * time.Minute),
			IsFromMe:       true,
			HandleID:       nil,
			HasAttachments: false,
		},
		{
			ID:             3,
			GUID:           "MSG-003-GUID-RECEIVED",
			Text:           &text3,
			Date:           int64(baseTime.Add(10 * time.Minute).UnixNano()),
			FormattedDate:  baseTime.Add(10 * time.Minute),
			IsFromMe:       false,
			HandleID:       &handleID2,
			HasAttachments: false,
		},
		{
			ID:             4,
			GUID:           "MSG-004-GUID-SENT",
			Text:           &text4,
			Date:           int64(baseTime.Add(15 * time.Minute).UnixNano()),
			FormattedDate:  baseTime.Add(15 * time.Minute),
			IsFromMe:       true,
			HandleID:       nil,
			HasAttachments: false,
		},
		{
			ID:             5,
			GUID:           "MSG-005-GUID-RECEIVED",
			Text:           &text5,
			Date:           int64(baseTime.Add(20 * time.Minute).UnixNano()),
			FormattedDate:  baseTime.Add(20 * time.Minute),
			IsFromMe:       false,
			HandleID:       &handleID1,
			HasAttachments: false,
		},
	}

	// Mock reactions
	reactions := map[string][]models.Reaction{
		"MSG-002-GUID-SENT": {
			{
				Type:          2001, // Like/Thumbs up
				SenderName:    "+14039090086",
				Timestamp:     baseTime.Add(6 * time.Minute),
				ReactionEmoji: "{\\emojifont\\symbol{\"1F44D}}",
			},
		},
		"MSG-003-GUID-RECEIVED": {
			{
				Type:          2000, // Love/Heart
				SenderName:    "Me",
				Timestamp:     baseTime.Add(11 * time.Minute),
				ReactionEmoji: "{\\emojifont\\symbol{\"2764}}",
			},
			{
				Type:          2003, // Laugh/Crying laugh emoji
				SenderName:    "+14039090086",
				Timestamp:     baseTime.Add(12 * time.Minute),
				ReactionEmoji: "{\\emojifont\\symbol{\"1F602}}",
			},
		},
	}

	return messages, handles, reactions
}

// TestGenerateBook tests the complete book generation flow
func TestGenerateBook(t *testing.T) {
	// Skip if we can't find template files
	if skipWithoutTemplates(t) {
		return
	}

	config := &models.BookConfig{
		Title:        "Test iMessage Book",
		Author:       "Test Author",
		IncludeImages: false, // Skip images for this basic test
		IncludePreviews: false, // Skip URL previews for this basic test
		OutputPath:   "test_book.md",
		PageWidth:    "5.5in",
		PageHeight:   "8.5in",
		TemplateDir:  "../templates/tex",
	}

	generator := New(config, nil) // Pass nil for db since we're not using URL processing

	messages, handles, reactions := getMockData()

	result := generator.GenerateBook(messages, handles, reactions)

	// Basic checks
	if !strings.Contains(result, "Test iMessage Book") {
		t.Error("Book should contain the title")
	}

	if !strings.Contains(result, "Hey, how's the fishing going?") {
		t.Error("Book should contain message content")
	}

	if !strings.Contains(result, "September 2025") {
		t.Error("Book should contain date headers")
	}

	if !strings.Contains(result, "\\begin{flushright}") {
		t.Error("Book should contain sent message formatting")
	}

	if !strings.Contains(result, "\\textbf{ +16047904258 }") {
		t.Error("Book should contain received message formatting with sender")
	}

	t.Logf("Generated book length: %d characters", len(result))
}

// TestReceivedMessageReactions tests received messages with reactions
func TestReceivedMessageReactions(t *testing.T) {
	// Skip if we can't find template files
	if skipWithoutTemplates(t) {
		return
	}

	config := &models.BookConfig{
		Title:           "Test Book",
		Author:          "Test Author",
		IncludeImages:   false,
		IncludePreviews: false,
		OutputPath:      "test_book.md",
		PageWidth:       "5.5in",
		PageHeight:      "8.5in",
		TemplateDir:     "../templates/tex",
	}

	generator := New(config, nil)

	messages, handles, reactions := getMockData()

	result := generator.GenerateBook(messages, handles, reactions)

	// Look for received message with multiple reactions
	if !strings.Contains(result, "emojifont\\symbol") {
		t.Error("Should contain emoji reactions")
	}

	t.Logf("Generated result contains reactions: %t", strings.Contains(result, "emojifont"))
}

// TestSentMessageReactions tests sent messages with reactions
func TestSentMessageReactions(t *testing.T) {
	// Skip if we can't find template files
	if skipWithoutTemplates(t) {
		return
	}

	config := &models.BookConfig{
		Title:           "Test Book",
		Author:          "Test Author",
		IncludeImages:   false,
		IncludePreviews: false,
		OutputPath:      "test_book.md",
		PageWidth:       "5.5in",
		PageHeight:      "8.5in",
		TemplateDir:     "../templates/tex",
	}

	generator := New(config, nil)

	messages, handles, reactions := getMockData()

	result := generator.GenerateBook(messages, handles, reactions)

	// Look for sent message with reaction
	if !strings.Contains(result, "\\begin{flushright}") {
		t.Error("Should contain right-aligned sent messages")
	}

	t.Logf("Generated result contains sent message formatting: %t", strings.Contains(result, "flushright"))
}

// TestMessageBubbleGeneration tests individual message bubble generation
func TestMessageBubbleGeneration(t *testing.T) {
	// Skip if we can't find template files
	if skipWithoutTemplates(t) {
		return
	}

	config := &models.BookConfig{
		Title:           "Test Book",
		Author:          "Test Author",
		IncludeImages:   false,
		IncludePreviews: false,
		OutputPath:      "test_book.md",
		PageWidth:       "5.5in",
		PageHeight:      "8.5in",
	}

	config.TemplateDir = "../templates/tex"
	generator := New(config, nil)

	_, _, reactions := getMockData()

	// Test received message with reactions
	var builder strings.Builder
	messageReactions := reactions["MSG-003-GUID-RECEIVED"]

	generator.writeReceivedMessageBubble(
		&builder,
		"That's awesome! 🎣",
		"10:10 AM",
		"+16047904258",
		true,
		true,
		messageReactions,
		nil, // No URL thumbnails for this test
	)

	result := builder.String()
	t.Logf("Generated received message bubble:\n%s", result)

	// Check structure
	if !strings.Contains(result, "\\begin{tabular}[t]") {
		t.Error("Should contain tabular structure")
	}

	if !strings.Contains(result, "+16047904258") {
		t.Error("Should contain sender name")
	}

	if !strings.Contains(result, "That's awesome! 🎣") {
		t.Error("Should contain message text")
	}

	if !strings.Contains(result, "emojifont\\symbol") {
		t.Error("Should contain reactions")
	}
}

// TestSentMessageBubbleGeneration tests sent message bubble generation
func TestSentMessageBubbleGeneration(t *testing.T) {
	// Skip if we can't find template files
	if skipWithoutTemplates(t) {
		return
	}

	config := &models.BookConfig{
		Title:           "Test Book",
		Author:          "Test Author",
		IncludeImages:   false,
		IncludePreviews: false,
		OutputPath:      "test_book.md",
		PageWidth:       "5.5in",
		PageHeight:      "8.5in",
		TemplateDir:     "../templates/tex",
	}

	generator := New(config, nil)

	_, _, reactions := getMockData()

	// Test sent message with reaction
	var builder strings.Builder
	messageReactions := reactions["MSG-002-GUID-SENT"]

	generator.writeSentMessageBubble(
		&builder,
		"Great! Caught a nice salmon today",
		"10:05 AM",
		messageReactions,
		nil, // No URL thumbnails for this test
	)

	result := builder.String()
	t.Logf("Generated sent message bubble:\n%s", result)

	// Check structure
	if !strings.Contains(result, "\\begin{flushright}") {
		t.Error("Should be right-aligned")
	}

	if !strings.Contains(result, "Great! Caught a nice salmon today") {
		t.Error("Should contain message text")
	}

	if !strings.Contains(result, "10:05 AM") {
		t.Error("Should contain timestamp")
	}

	if !strings.Contains(result, "emojifont\\symbol") {
		t.Error("Should contain reaction")
	}
}

// TestObjectReplacementCharacterRemoval tests that U+FFFC is removed from messages
func TestObjectReplacementCharacterRemoval(t *testing.T) {
	// Skip if we can't find template files
	if skipWithoutTemplates(t) {
		return
	}

	config := &models.BookConfig{
		Title:           "Test Book",
		Author:          "Test Author",
		IncludeImages:   false,
		IncludePreviews: false,
		OutputPath:      "test_book.md",
		PageWidth:       "5.5in",
		PageHeight:      "8.5in",
		TemplateDir:     "../templates/tex",
	}

	generator := New(config, nil)

	// Test sent message with object replacement character
	var sentBuilder strings.Builder
	generator.writeSentMessageBubble(
		&sentBuilder,
		"Check this out \uFFFC isn't that cool?",
		"10:00 AM",
		nil,
		nil,
	)

	sentResult := sentBuilder.String()
	if strings.Contains(sentResult, "\uFFFC") {
		t.Error("Sent message should not contain object replacement character (U+FFFC)")
	}
	if !strings.Contains(sentResult, "Check this out") {
		t.Error("Sent message should contain the text before the character")
	}
	if !strings.Contains(sentResult, "isn't that cool?") {
		t.Error("Sent message should contain the text after the character")
	}

	// Test received message with object replacement character
	var receivedBuilder strings.Builder
	generator.writeReceivedMessageBubble(
		&receivedBuilder,
		"Look at this ￼ amazing thing",
		"10:05 AM",
		"John Doe",
		true,
		true,
		nil,
		nil,
	)

	receivedResult := receivedBuilder.String()
	if strings.Contains(receivedResult, "\uFFFC") || strings.Contains(receivedResult, "￼") {
		t.Error("Received message should not contain object replacement character")
	}
	if !strings.Contains(receivedResult, "Look at this") {
		t.Error("Received message should contain the text before the character")
	}
	if !strings.Contains(receivedResult, "amazing thing") {
		t.Error("Received message should contain the text after the character")
	}

	t.Logf("Successfully removed object replacement character from both sent and received messages")
}

// Helper function to skip tests when template files aren't available
func skipWithoutTemplates(t *testing.T) bool {
	templatePath := "../templates/tex/sent-message.tex"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Skip("Template files not available - running in extracted test environment")
		return true
	}
	return false
}