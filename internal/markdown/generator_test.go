package markdown

import (
	"strings"
	"testing"
	"time"

	"imessages-book/internal/models"
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
			ID:            1,
			GUID:          "MSG-001-GUID-RECEIVED",
			Text:          &text1,
			Date:          int64(baseTime.UnixNano()),
			FormattedDate: baseTime,
			IsFromMe:      false,
			HandleID:      &handleID1,
			HasAttachments: false,
		},
		{
			ID:            2,
			GUID:          "MSG-002-GUID-SENT",
			Text:          &text2,
			Date:          int64(baseTime.Add(5 * time.Minute).UnixNano()),
			FormattedDate: baseTime.Add(5 * time.Minute),
			IsFromMe:      true,
			HandleID:      nil,
			HasAttachments: false,
		},
		{
			ID:            3,
			GUID:          "MSG-003-GUID-RECEIVED",
			Text:          &text3,
			Date:          int64(baseTime.Add(10 * time.Minute).UnixNano()),
			FormattedDate: baseTime.Add(10 * time.Minute),
			IsFromMe:      false,
			HandleID:      &handleID2,
			HasAttachments: false,
		},
		{
			ID:            4,
			GUID:          "MSG-004-GUID-SENT",
			Text:          &text4,
			Date:          int64(baseTime.Add(15 * time.Minute).UnixNano()),
			FormattedDate: baseTime.Add(15 * time.Minute),
			IsFromMe:      true,
			HandleID:      nil,
			HasAttachments: false,
		},
		{
			ID:            5,
			GUID:          "MSG-005-GUID-RECEIVED",
			Text:          &text5,
			Date:          int64(baseTime.Add(20 * time.Minute).UnixNano()),
			FormattedDate: baseTime.Add(20 * time.Minute),
			IsFromMe:      false,
			HandleID:      &handleID1,
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
				Type:          2000, // Love
				SenderName:    "Me",
				Timestamp:     baseTime.Add(11 * time.Minute),
				ReactionEmoji: "{\\emojifont\\symbol{\"2764}}",
			},
			{
				Type:          2003, // Laugh
				SenderName:    "+14039090086",
				Timestamp:     baseTime.Add(12 * time.Minute),
				ReactionEmoji: "{\\emojifont\\symbol{\"1F602}}",
			},
		},
		"MSG-004-GUID-SENT": {
			{
				Type:          2001, // Like
				SenderName:    "+16047904258",
				Timestamp:     baseTime.Add(16 * time.Minute),
				ReactionEmoji: "{\\emojifont\\symbol{\"1F44D}}",
			},
		},
	}

	return messages, handles, reactions
}

// Test basic markdown generation
func TestGenerateBook(t *testing.T) {
	config := &models.BookConfig{
		Title:      "Test Messages",
		Author:     "Test Author",
		OutputPath: "test_book.md",
		PageWidth:  "5.5in",
		PageHeight: "8.5in",
	}

	generator := New(config)
	messages, handles, reactions := getMockData()

	result := generator.GenerateBook(messages, handles, reactions)

	// Basic checks
	if !strings.Contains(result, "Test Messages") {
		t.Error("Generated book should contain title")
	}

	if !strings.Contains(result, "Hey, how's the fishing going?") {
		t.Error("Generated book should contain message text")
	}

	if !strings.Contains(result, "emojifont") {
		t.Error("Generated book should contain reaction emojis")
	}
}

// Test reaction positioning for received messages
func TestReceivedMessageReactions(t *testing.T) {
	config := &models.BookConfig{
		Title:      "Test Messages",
		Author:     "Test Author",
		OutputPath: "test_book.md",
		PageWidth:  "5.5in",
		PageHeight: "8.5in",
	}

	generator := New(config)
	messages, handles, reactions := getMockData()

	result := generator.GenerateBook(messages, handles, reactions)

	// Look for received message with multiple reactions
	// MSG-003-GUID-RECEIVED has 2 reactions
	if !strings.Contains(result, "That's awesome! 🎣") {
		t.Error("Should contain the message with reactions")
	}

	// Check for tabular structure with reactions
	if !strings.Contains(result, "\\begin{tabular}[t]") {
		t.Error("Should use top-aligned tabular structure")
	}

	// Check for multiple reactions with line breaks
	if !strings.Contains(result, "\\\\") {
		t.Error("Should have line breaks between multiple reactions")
	}

	// Check for dark grey color
	if !strings.Contains(result, "\\textcolor{darkgray}") {
		t.Error("Reactions should use dark grey color")
	}
}

// Test reaction positioning for sent messages
func TestSentMessageReactions(t *testing.T) {
	config := &models.BookConfig{
		Title:      "Test Messages",
		Author:     "Test Author",
		OutputPath: "test_book.md",
		PageWidth:  "5.5in",
		PageHeight: "8.5in",
	}

	generator := New(config)
	messages, handles, reactions := getMockData()

	result := generator.GenerateBook(messages, handles, reactions)

	// Look for sent message with reaction
	// MSG-002-GUID-SENT and MSG-004-GUID-SENT have reactions
	if !strings.Contains(result, "Great! Caught a nice salmon today") {
		t.Error("Should contain the sent message with reactions")
	}

	// Check for flushright wrapper
	if !strings.Contains(result, "\\begin{flushright}") {
		t.Error("Sent messages should be wrapped in flushright")
	}

	// Check for tabular structure
	if !strings.Contains(result, "\\begin{tabular}[t]") {
		t.Error("Should use top-aligned tabular structure")
	}
}

// Test template execution for individual messages
func TestMessageBubbleGeneration(t *testing.T) {
	config := &models.BookConfig{
		Title:      "Test Messages",
		Author:     "Test Author",
		OutputPath: "test_book.md",
		PageWidth:  "5.5in",
		PageHeight: "8.5in",
	}

	generator := New(config)
	_, _, reactions := getMockData()

	// Test received message with reactions
	var builder strings.Builder
	messageReactions := reactions["MSG-003-GUID-RECEIVED"]

	generator.writeReceivedMessageBubble(
		&builder,
		"That's awesome! 🎣",
		"10:10 AM",
		"+16047904258",
		true,  // showSender
		true,  // showTimestamp
		messageReactions,
	)

	result := builder.String()

	// Check structure
	if !strings.Contains(result, "\\textbf{ +16047904258 }") {
		t.Error("Should show sender name")
	}

	if !strings.Contains(result, "10:10 AM") {
		t.Error("Should show timestamp")
	}

	if !strings.Contains(result, "That's awesome! 🎣") {
		t.Error("Should show message text")
	}

	if !strings.Contains(result, "emojifont\\symbol") {
		t.Error("Should show reactions")
	}

	if !strings.Contains(result, "\\\\") {
		t.Error("Should have line breaks between multiple reactions")
	}

	t.Logf("Generated received message bubble:\n%s", result)
}

// Test sent message bubble
func TestSentMessageBubbleGeneration(t *testing.T) {
	config := &models.BookConfig{
		Title:      "Test Messages",
		Author:     "Test Author",
		OutputPath: "test_book.md",
		PageWidth:  "5.5in",
		PageHeight: "8.5in",
	}

	generator := New(config)
	_, _, reactions := getMockData()

	// Test sent message with reaction
	var builder strings.Builder
	messageReactions := reactions["MSG-002-GUID-SENT"]

	generator.writeSentMessageBubble(
		&builder,
		"Great! Caught a nice salmon today",
		"10:05 AM",
		messageReactions,
	)

	result := builder.String()

	// Check structure
	if !strings.Contains(result, "\\begin{flushright}") {
		t.Error("Should be wrapped in flushright")
	}

	if !strings.Contains(result, "Great! Caught a nice salmon today") {
		t.Error("Should show message text")
	}

	if !strings.Contains(result, "10:05 AM") {
		t.Error("Should show timestamp")
	}

	if !strings.Contains(result, "emojifont\\symbol") {
		t.Error("Should show reactions")
	}

	t.Logf("Generated sent message bubble:\n%s", result)
}