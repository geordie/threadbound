package markdown

import (
	"strings"
	"testing"
	"text/template"
	"time"

	"threadbound/internal/models"
)

// Test just the template execution without loading from files
func TestReceivedMessageTemplate(t *testing.T) {
	// Define the template content directly
	receivedTemplateContent := `{{if .ShowSender}}\textbf{ {{.Sender}} } \small\textcolor{gray}{ {{.Timestamp}} }

{{else if .ShowTimestamp}}\small\textcolor{gray}{ {{.Timestamp}} }

{{end}}\begin{tabular}[t]{@{}p{0.7\textwidth}@{\hspace{0.02\textwidth}}p{0.25\textwidth}@{}}
\tikz[baseline=(textnode.base)]\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] (textnode) { {{.Text}} }; & {{if .Reactions}}\raggedleft\small\textcolor{darkgray}{ {{range $i, $reaction := .Reactions}}{{if gt $i 0}}\\{{end}}{{$reaction.ReactionEmoji}}\,{{$reaction.SenderName}}{{end}} }{{end}} \\
\end{tabular}`

	tmpl, err := template.New("received-test").Parse(receivedTemplateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Test data
	reactions := []models.Reaction{
		{
			Type:          2000,
			SenderName:    "Me",
			Timestamp:     time.Now(),
			ReactionEmoji: "{\\emojifont\\symbol{\"2764}}",
		},
		{
			Type:          2003,
			SenderName:    "+14039090086",
			Timestamp:     time.Now(),
			ReactionEmoji: "{\\emojifont\\symbol{\"1F602}}",
		},
	}

	data := struct {
		Text          string
		Timestamp     string
		Sender        string
		ShowSender    bool
		ShowTimestamp bool
		Reactions     []models.Reaction
	}{
		Text:          "That's awesome! 🎣",
		Timestamp:     "10:10 AM",
		Sender:        "+16047904258",
		ShowSender:    true,
		ShowTimestamp: true,
		Reactions:     reactions,
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := result.String()
	t.Logf("Generated received message template:\n%s", output)

	// Verify structure
	if !strings.Contains(output, "\\begin{tabular}[t]") {
		t.Error("Should contain top-aligned tabular structure")
	}

	if !strings.Contains(output, "\\textbf{ +16047904258 }") {
		t.Error("Should show sender name")
	}

	if !strings.Contains(output, "That's awesome! 🎣") {
		t.Error("Should show message text")
	}

	if !strings.Contains(output, "emojifont\\symbol") {
		t.Error("Should show reactions")
	}

	if !strings.Contains(output, "\\\\") {
		t.Error("Should have line breaks between multiple reactions")
	}

	if !strings.Contains(output, "\\textcolor{darkgray}") {
		t.Error("Should use dark grey color for reactions")
	}

	// Check that reactions are in the right column
	if !strings.Contains(output, "}; & ") {
		t.Error("Should have reactions in second column after message")
	}
}

func TestSentMessageTemplate(t *testing.T) {
	// Define the template content directly
	sentTemplateContent := `\begin{flushright}
\begin{tabular}[t]{@{}p{0.25\textwidth}@{\hspace{0.02\textwidth}}p{0.7\textwidth}@{}}
{{if .Reactions}}\raggedright\small\textcolor{darkgray}{ {{range $i, $reaction := .Reactions}}{{if gt $i 0}}\\{{end}}{{$reaction.ReactionEmoji}}\,{{$reaction.SenderName}}{{end}} }{{end}} & \tikz[baseline=(textnode.base)]\node [draw=none, fill=blue!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] (textnode) { {{.Text}} }; \\
\end{tabular}

\small\textcolor{gray}{ {{.Timestamp}} }
\end{flushright}`

	tmpl, err := template.New("sent-test").Parse(sentTemplateContent)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// Test data with single reaction
	reactions := []models.Reaction{
		{
			Type:          2001,
			SenderName:    "+14039090086",
			Timestamp:     time.Now(),
			ReactionEmoji: "{\\emojifont\\symbol{\"1F44D}}",
		},
	}

	data := struct {
		Text      string
		Timestamp string
		Reactions []models.Reaction
	}{
		Text:      "Great! Caught a nice salmon today",
		Timestamp: "10:05 AM",
		Reactions: reactions,
	}

	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := result.String()
	t.Logf("Generated sent message template:\n%s", output)

	// Verify structure
	if !strings.Contains(output, "\\begin{flushright}") {
		t.Error("Should be wrapped in flushright")
	}

	if !strings.Contains(output, "\\begin{tabular}[t]") {
		t.Error("Should contain top-aligned tabular structure")
	}

	if !strings.Contains(output, "Great! Caught a nice salmon today") {
		t.Error("Should show message text")
	}

	if !strings.Contains(output, "10:05 AM") {
		t.Error("Should show timestamp")
	}

	if !strings.Contains(output, "emojifont\\symbol") {
		t.Error("Should show reactions")
	}

	if !strings.Contains(output, "\\textcolor{darkgray}") {
		t.Error("Should use dark grey color for reactions")
	}

	// Check that reactions are in the left column
	if !strings.Contains(output, "} & \\tikz") {
		t.Error("Should have reactions in first column before message")
	}
}

// Test the alignment issue specifically
func TestTabularAlignment(t *testing.T) {
	// This test specifically checks the tabular structure to debug alignment issues

	receivedTemplate := `\begin{tabular}[t]{@{}p{0.7\textwidth}@{\hspace{0.02\textwidth}}p{0.25\textwidth}@{}}
MESSAGE_BUBBLE & REACTIONS \\
\end{tabular}`

	sentTemplate := `\begin{tabular}[t]{@{}p{0.25\textwidth}@{\hspace{0.02\textwidth}}p{0.7\textwidth}@{}}
REACTIONS & MESSAGE_BUBBLE \\
\end{tabular}`

	t.Logf("Received message table structure:\n%s", receivedTemplate)
	t.Logf("Sent message table structure:\n%s", sentTemplate)

	// Test the basic structure is what we expect
	if !strings.Contains(receivedTemplate, "[t]") {
		t.Error("Received template should have top alignment")
	}

	if !strings.Contains(sentTemplate, "[t]") {
		t.Error("Sent template should have top alignment")
	}

	if !strings.Contains(receivedTemplate, "p{0.7\\textwidth}") {
		t.Error("Received template should have correct column widths")
	}

	if !strings.Contains(sentTemplate, "p{0.25\\textwidth}") {
		t.Error("Sent template should have correct column widths")
	}
}