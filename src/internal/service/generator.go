package service

import (
	"fmt"

	"threadbound/internal/book"
	"threadbound/internal/models"
)

// GeneratorService handles book generation logic
type GeneratorService struct {
	config *models.BookConfig
}

// NewGeneratorService creates a new generator service
func NewGeneratorService(config *models.BookConfig) *GeneratorService {
	return &GeneratorService{
		config: config,
	}
}

// GenerateResult contains the result of a generation operation
type GenerateResult struct {
	OutputPath string
	Stats      *models.BookStats
}

// Generate executes the book generation process
func (s *GeneratorService) Generate() (*GenerateResult, error) {
	// Create book builder
	builder, err := book.New(s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create builder: %w", err)
	}
	defer builder.Close()

	// Get statistics
	stats, err := builder.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Generate the book
	err = builder.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate book: %w", err)
	}

	return &GenerateResult{
		OutputPath: s.config.OutputPath,
		Stats:      stats,
	}, nil
}

// BuildPDF executes the PDF build process
func (s *GeneratorService) BuildPDF() (*models.PDFInfo, error) {
	// Create PDF builder
	pdfBuilder := book.NewPDFBuilder(s.config)

	// Generate output filename
	outputPDF := "book.pdf"
	if s.config.OutputPath != "book.tex" && len(s.config.OutputPath) > 4 && s.config.OutputPath[len(s.config.OutputPath)-4:] == ".tex" {
		// Replace .tex extension with .pdf
		outputPDF = s.config.OutputPath[:len(s.config.OutputPath)-4] + ".pdf"
	}

	// Build the PDF
	err := pdfBuilder.BuildPDF(s.config.OutputPath, outputPDF)
	if err != nil {
		return nil, fmt.Errorf("failed to build PDF: %w", err)
	}

	// Get PDF info
	info, err := pdfBuilder.GetPDFInfo(outputPDF)
	if err != nil {
		return nil, fmt.Errorf("failed to get PDF info: %w", err)
	}

	return info, nil
}

// GetStats returns statistics about the messages without generating
func (s *GeneratorService) GetStats() (*models.BookStats, error) {
	builder, err := book.New(s.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create builder: %w", err)
	}
	defer builder.Close()

	return builder.GetStats()
}
