package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"threadbound/internal/book"
	"threadbound/internal/models"
)

var config models.BookConfig
var configFile string

var rootCmd = &cobra.Command{
	Use:   "threadbound",
	Short: "Convert iMessages database to a book",
	Long: `A tool to extract iMessages from a SQLite database and convert them
into a formatted book using XeLaTeX.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate TeX from iMessages database",
	Long:  `Extract messages from the SQLite database and generate a TeX file`,
	PreRunE: loadConfig,
	RunE:  runGenerate,
}

var buildCmd = &cobra.Command{
	Use:   "build-pdf",
	Short: "Build PDF from TeX using XeLaTeX",
	Long:  `Convert the generated TeX to PDF using XeLaTeX`,
	PreRunE: loadConfig,
	RunE:  runBuildPDF,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to config file (YAML format)")

	// Initialize config with defaults
	defaultConfig := models.GetDefaultConfig()
	config = *defaultConfig

	// Generate command flags
	generateCmd.Flags().StringVar(&config.DatabasePath, "db", "chat.db", "Path to iMessages database")
	generateCmd.Flags().StringVar(&config.AttachmentsPath, "attachments", "Attachments", "Path to attachments directory")
	generateCmd.Flags().StringVar(&config.OutputPath, "output", "book.tex", "Output TeX file")
	generateCmd.Flags().StringVar(&config.Title, "title", "Our Messages", "Book title")
	generateCmd.Flags().StringVar(&config.Author, "author", "", "Book author")
	generateCmd.Flags().StringVar(&config.PageWidth, "page-width", "5.5in", "Page width")
	generateCmd.Flags().StringVar(&config.PageHeight, "page-height", "8.5in", "Page height")
	generateCmd.Flags().BoolVar(&config.IncludeImages, "include-images", true, "Include images in output")

	// Always enable URL previews
	config.IncludePreviews = true

	// Build command flags
	buildCmd.Flags().StringVar(&config.OutputPath, "input", "book.tex", "Input TeX file")
	buildCmd.Flags().StringVar(&config.TemplateDir, "template-dir", "internal/templates/tex", "Template directory")
	buildCmd.Flags().StringVar(&config.PageWidth, "page-width", "5.5in", "Page width")
	buildCmd.Flags().StringVar(&config.PageHeight, "page-height", "8.5in", "Page height")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(buildCmd)
}

// loadConfig loads configuration from file if specified, otherwise uses defaults and flags
func loadConfig(cmd *cobra.Command, args []string) error {
	if configFile != "" {
		fileConfig, err := models.LoadConfigFromFile(configFile)
		if err != nil {
			return err
		}

		// Merge file config with command-line flags
		// Command-line flags take precedence over config file
		if !cmd.Flags().Changed("title") && fileConfig.Title != "" {
			config.Title = fileConfig.Title
		}
		if !cmd.Flags().Changed("author") && fileConfig.Author != "" {
			config.Author = fileConfig.Author
		}
		if !cmd.Flags().Changed("db") && fileConfig.DatabasePath != "" {
			config.DatabasePath = fileConfig.DatabasePath
		}
		if !cmd.Flags().Changed("attachments") && fileConfig.AttachmentsPath != "" {
			config.AttachmentsPath = fileConfig.AttachmentsPath
		}
		if !cmd.Flags().Changed("output") && fileConfig.OutputPath != "" {
			config.OutputPath = fileConfig.OutputPath
		}
		if !cmd.Flags().Changed("template-dir") && fileConfig.TemplateDir != "" {
			config.TemplateDir = fileConfig.TemplateDir
		}
		if !cmd.Flags().Changed("include-images") {
			config.IncludeImages = fileConfig.IncludeImages
		}
		if !cmd.Flags().Changed("page-width") && fileConfig.PageWidth != "" {
			config.PageWidth = fileConfig.PageWidth
		}
		if !cmd.Flags().Changed("page-height") && fileConfig.PageHeight != "" {
			config.PageHeight = fileConfig.PageHeight
		}

		// Merge contact names from config file
		if fileConfig.ContactNames != nil {
			config.ContactNames = fileConfig.ContactNames
		}

		// IncludePreviews is always enabled for now
		config.IncludePreviews = true
	}
	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	fmt.Printf("📚 iMessages Book Generator\n")
	fmt.Printf("Database: %s\n", config.DatabasePath)
	fmt.Printf("Output: %s\n", config.OutputPath)
	fmt.Printf("Title: %s\n", config.Title)
	fmt.Println()

	// Create book builder
	builder, err := book.New(&config)
	if err != nil {
		return err
	}
	defer builder.Close()

	// Show statistics first
	stats, err := builder.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("📊 Book Statistics:\n")
	fmt.Printf("   Messages: %d (%d with text)\n", stats.TotalMessages, stats.TextMessages)
	fmt.Printf("   Contacts: %d\n", stats.TotalContacts)
	fmt.Printf("   Attachments: %d\n", stats.AttachmentCount)
	if !stats.StartDate.IsZero() && !stats.EndDate.IsZero() {
		fmt.Printf("   Date Range: %s to %s\n",
			stats.StartDate.Format("Jan 2, 2006"),
			stats.EndDate.Format("Jan 2, 2006"))
	}
	fmt.Println()

	// Generate the book
	return builder.Generate()
}


func runBuildPDF(cmd *cobra.Command, args []string) error {
	fmt.Printf("📚 iMessages PDF Builder\n")
	fmt.Printf("Input: %s\n", config.OutputPath)
	fmt.Printf("Template Dir: %s\n", config.TemplateDir)
	fmt.Printf("Page Size: %s x %s\n", config.PageWidth, config.PageHeight)
	fmt.Println()

	// Create PDF builder
	pdfBuilder := book.NewPDFBuilder(&config)

	// Generate output filename
	outputPDF := "book.pdf"
	if config.OutputPath != "book.tex" && len(config.OutputPath) > 4 && config.OutputPath[len(config.OutputPath)-4:] == ".tex" {
		// Replace .tex extension with .pdf
		outputPDF = config.OutputPath[:len(config.OutputPath)-4] + ".pdf"
	}

	// Build the PDF
	err := pdfBuilder.BuildPDF(config.OutputPath, outputPDF)
	if err != nil {
		return err
	}

	// Show PDF info
	info, err := pdfBuilder.GetPDFInfo(outputPDF)
	if err != nil {
		fmt.Printf("⚠️  Could not get PDF info: %v\n", err)
	} else {
		fmt.Printf("📊 PDF Info:\n")
		fmt.Printf("   File: %s\n", info.FilePath)
		fmt.Printf("   Size: %d bytes (%.2f MB)\n", info.FileSize, float64(info.FileSize)/(1024*1024))
		fmt.Printf("   Dimensions: %s x %s\n", info.PageWidth, info.PageHeight)
	}

	// Suggest preview command
	previewCmd := pdfBuilder.PreviewCommand(outputPDF)
	fmt.Printf("\n📖 To preview: %s\n", previewCmd)

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}