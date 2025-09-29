package texgen

import (
	"strings"
	"testing"
)

// Test different alignment approaches
func TestAlignmentApproaches(t *testing.T) {
	// Approach 1: Current - table with [t] alignment
	approach1 := `\begin{tabular}[t]{@{}p{0.7\textwidth}@{\hspace{0.02\textwidth}}p{0.25\textwidth}@{}}
\tikz\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] { Message text }; & \raggedleft\small\textcolor{darkgray}{ Reaction } \\
\end{tabular}`

	// Approach 2: Use array package with >{} column specifications
	approach2 := `\begin{tabular}{@{}>{\raggedright\arraybackslash}p{0.7\textwidth}@{\hspace{0.02\textwidth}}>{\raggedleft\arraybackslash}p{0.25\textwidth}@{}}
\tikz\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] { Message text }; & \small\textcolor{darkgray}{ Reaction } \\
\end{tabular}`

	// Approach 3: Use m{} columns for middle alignment
	approach3 := `\begin{tabular}{@{}m{0.7\textwidth}@{\hspace{0.02\textwidth}}m{0.25\textwidth}@{}}
\tikz\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] { Message text }; & \raggedleft\small\textcolor{darkgray}{ Reaction } \\
\end{tabular}`

	// Approach 4: Use manual vertical positioning with \raisebox
	approach4 := `\begin{tabular}[t]{@{}p{0.7\textwidth}@{\hspace{0.02\textwidth}}p{0.25\textwidth}@{}}
\tikz\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] { Message text }; & \raisebox{\dimexpr-\height+\ht\strutbox\relax}{\raggedleft\small\textcolor{darkgray}{ Reaction }} \\
\end{tabular}`

	// Approach 5: Use longtabu or other packages
	approach5 := `\noindent\begin{minipage}[t]{0.7\textwidth}
\tikz\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] { Message text };
\end{minipage}\hfill\begin{minipage}[t]{0.25\textwidth}
\vspace{0pt}
\raggedleft\small\textcolor{darkgray}{ Reaction }
\end{minipage}`

	t.Logf("Approach 1 (Current [t] table):\n%s\n", approach1)
	t.Logf("Approach 2 (Array package):\n%s\n", approach2)
	t.Logf("Approach 3 (m{} columns):\n%s\n", approach3)
	t.Logf("Approach 4 (raisebox):\n%s\n", approach4)
	t.Logf("Approach 5 (minipage with vspace{0pt}):\n%s\n", approach5)

	// All should have basic structure
	approaches := []string{approach1, approach2, approach3, approach4, approach5}
	for i, approach := range approaches {
		if !strings.Contains(approach, "Message text") {
			t.Errorf("Approach %d should contain message", i+1)
		}
		if !strings.Contains(approach, "Reaction") {
			t.Errorf("Approach %d should contain reaction", i+1)
		}
	}
}

// Test the specific issue: tikz node with inner sep creates vertical offset
func TestTikzNodeAlignment(t *testing.T) {
	// The problem might be that tikz node with inner sep=8pt creates padding
	// that shifts the text down, but we want reactions to align with the text, not the top of the padding

	// Option 1: Align reactions with tikz node baseline
	option1 := `\begin{tabular}[t]{@{}p{0.7\textwidth}@{\hspace{0.02\textwidth}}p{0.25\textwidth}@{}}
\tikz[baseline=(textnode.base)]\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] (textnode) { Message text }; & \raggedleft\small\textcolor{darkgray}{ Reaction } \\
\end{tabular}`

	// Option 2: Move reactions down to match inner sep
	option2 := `\begin{tabular}[t]{@{}p{0.7\textwidth}@{\hspace{0.02\textwidth}}p{0.25\textwidth}@{}}
\tikz\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt] { Message text }; & \raisebox{-8pt}{\raggedleft\small\textcolor{darkgray}{ Reaction }} \\
\end{tabular}`

	// Option 3: Zero out tikz node padding effects
	option3 := `\begin{tabular}[t]{@{}p{0.7\textwidth}@{\hspace{0.02\textwidth}}p{0.25\textwidth}@{}}
\tikz[baseline]\node [draw=none, fill=gray!20, rounded corners=4pt, text width=0.7\textwidth, align=left, inner sep=8pt, text depth=0pt] { Message text }; & \raggedleft\small\textcolor{darkgray}{ Reaction } \\
\end{tabular}`

	t.Logf("Option 1 (tikz baseline):\n%s\n", option1)
	t.Logf("Option 2 (raisebox -8pt):\n%s\n", option2)
	t.Logf("Option 3 (tikz baseline + text depth):\n%s\n", option3)
}