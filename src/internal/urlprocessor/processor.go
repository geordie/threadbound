package urlprocessor

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"threadbound/internal/models"
	"threadbound/internal/output"
)

// URLProcessor handles URL detection and preview extraction from iMessage database
type URLProcessor struct {
	config    *models.BookConfig
	cacheDir  string
	urlRegex  *regexp.Regexp
	db        *sql.DB
}

// URLThumbnail is an alias to output.URLThumbnail for backward compatibility
type URLThumbnail = output.URLThumbnail

// New creates a new URL processor
func New(config *models.BookConfig, db *sql.DB) *URLProcessor {
	// Create cache directory for URL thumbnails
	cacheDir := filepath.Join(config.AttachmentsPath, "url-thumbnails")
	os.MkdirAll(cacheDir, 0755)

	// Regex to match HTTP/HTTPS URLs
	urlRegex := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)

	return &URLProcessor{
		config:   config,
		cacheDir: cacheDir,
		urlRegex: urlRegex,
		db:       db,
	}
}

// FindURLsInText extracts all URLs from message text
func (p *URLProcessor) FindURLsInText(text string) []string {
	if text == "" {
		return nil
	}

	matches := p.urlRegex.FindAllString(text, -1)

	// Remove duplicates and clean URLs
	urlMap := make(map[string]bool)
	var urls []string

	for _, match := range matches {
		cleanURL := strings.TrimRight(match, ".,;!?)")
		if !urlMap[cleanURL] {
			urlMap[cleanURL] = true
			urls = append(urls, cleanURL)
		}
	}

	return urls
}

// ProcessMessageForURLPreviews extracts URL preview data from a message in the database
func (p *URLProcessor) ProcessMessageForURLPreviews(messageID int64) map[string]*URLThumbnail {
	results := make(map[string]*URLThumbnail)

	// Check if message has payload_data (rich link metadata)
	var payloadData []byte
	var messageText string
	err := p.db.QueryRow("SELECT text, payload_data FROM message WHERE ROWID = ?", messageID).Scan(&messageText, &payloadData)
	if err != nil || len(payloadData) == 0 {
		return results
	}

	// Find URLs in the message text
	urls := p.FindURLsInText(messageText)
	if len(urls) == 0 {
		return results
	}

	// Extract rich link metadata from payload_data
	metadata, err := p.extractRichLinkMetadata(payloadData, urls[0])
	if err != nil {
		return results
	}

	// Get attachments for this message
	attachments, err := p.getMessageAttachments(messageID)
	if err != nil {
		return results
	}

	// Process the first URL found (iMessage typically shows preview for first URL)
	if len(urls) > 0 {
		url := urls[0]
		thumbnail := p.createThumbnailFromMetadata(url, metadata, attachments)
		if thumbnail != nil {
			results[url] = thumbnail
		}
	}

	return results
}

// ProcessURL generates a thumbnail for a URL (fallback method)
func (p *URLProcessor) ProcessURL(urlStr string) *URLThumbnail {
	result := &URLThumbnail{
		URL:     urlStr,
		Success: false,
	}

	// Generate cache filename based on URL hash
	hash := fmt.Sprintf("%x", md5.Sum([]byte(urlStr)))
	thumbnailPath := filepath.Join(p.cacheDir, hash+".png")

	// Check if thumbnail already exists
	if _, err := os.Stat(thumbnailPath); err == nil {
		result.ThumbnailPath = thumbnailPath
		result.Success = true
		result.Title = p.extractDomainTitle(urlStr)
		return result
	}

	// Generate a simple domain card as fallback
	success := p.generateDomainCard(urlStr, thumbnailPath, result)

	result.Success = success
	if success {
		result.ThumbnailPath = thumbnailPath
	}

	return result
}

// RichLinkMetadata represents extracted metadata from iMessage rich links
type RichLinkMetadata struct {
	Title       string
	Summary     string
	SiteName    string
	ImageIndex  int
	IconIndex   int
	HasImage    bool
	HasIcon     bool
	ImageURL    string
	IconURL     string
}

// MessageAttachment represents an attachment linked to a message
type MessageAttachment struct {
	ID       int64
	GUID     string
	MIMEType string
	Filename string
	Data     []byte
}

// extractRichLinkMetadata parses the payload_data plist to extract metadata
func (p *URLProcessor) extractRichLinkMetadata(payloadData []byte, originalURL string) (*RichLinkMetadata, error) {
	// Write payload data to temporary file
	tmpFile := filepath.Join(os.TempDir(), "payload_data.plist")
	err := os.WriteFile(tmpFile, payloadData, 0644)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	// Use plutil to convert to readable format
	cmd := exec.Command("plutil", "-p", tmpFile)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Parse the output to extract metadata
	metadata := &RichLinkMetadata{}
	outputStr := string(output)

	// Extract title via UID reference
	if titleUID := extractPlistValue(outputStr, `"title" => <[^>]+>\{value = (\d+)\}`); titleUID != "" {
		if idx := parseInt(titleUID); idx >= 0 {
			pattern := fmt.Sprintf(`\s+%d => "([^"]+)"`, idx)
			if title := extractPlistValue(outputStr, pattern); title != "" {
				metadata.Title = title
			}
		}
	}

	// Extract summary via UID reference
	if summaryUID := extractPlistValue(outputStr, `"summary" => <[^>]+>\{value = (\d+)\}`); summaryUID != "" {
		if idx := parseInt(summaryUID); idx >= 0 {
			pattern := fmt.Sprintf(`\s+%d => "([^"]+)"`, idx)
			if summary := extractPlistValue(outputStr, pattern); summary != "" {
				metadata.Summary = summary
			}
		}
	}

	// Extract site name via UID reference
	if siteNameUID := extractPlistValue(outputStr, `"siteName" => <[^>]+>\{value = (\d+)\}`); siteNameUID != "" {
		if idx := parseInt(siteNameUID); idx >= 0 {
			pattern := fmt.Sprintf(`\s+%d => "([^"]+)"`, idx)
			if siteName := extractPlistValue(outputStr, pattern); siteName != "" {
				metadata.SiteName = siteName
			}
		}
	}

	// Check for image attachment substitute index
	if imageIndex := extractPlistValue(outputStr, `"richLinkImageAttachmentSubstituteIndex" => (\d+)`); imageIndex != "" {
		if idx := parseInt(imageIndex); idx >= 0 {
			metadata.ImageIndex = idx
			metadata.HasImage = true
		}
	}

	// Extract all URLs from the plist and categorize them
	allURLs := extractAllURLs(outputStr)

	// Categorize URLs by priority
	var previewURLs []string
	var iconURLs []string

	for _, url := range allURLs {
		if isPreviewImageURL(url) {
			previewURLs = append(previewURLs, url)
		} else if isIconURL(url) {
			iconURLs = append(iconURLs, url)
		}
	}

	// Use the highest priority preview image URL
	if len(previewURLs) > 0 {
		metadata.ImageURL = previewURLs[0]
		metadata.HasImage = true
		fmt.Printf("🖼️ Found preview image: %s\n", metadata.ImageURL)
	} else {
		// Try to reconstruct preview URLs for services that don't include them
		if reconstructedURL := p.reconstructPreviewURL(outputStr, originalURL); reconstructedURL != "" {
			metadata.ImageURL = reconstructedURL
			metadata.HasImage = true
			fmt.Printf("🔧 Reconstructed preview image: %s\n", metadata.ImageURL)
		}
	}

	// Use the first icon URL if available
	if len(iconURLs) > 0 {
		metadata.IconURL = iconURLs[0]
		metadata.HasIcon = true
		fmt.Printf("🔗 Found icon: %s\n", metadata.IconURL)
	}

	return metadata, nil
}

// getMessageAttachments retrieves attachments for a message
func (p *URLProcessor) getMessageAttachments(messageID int64) ([]MessageAttachment, error) {
	var attachments []MessageAttachment

	query := `
		SELECT a.ROWID, a.guid, a.mime_type, a.filename
		FROM attachment a
		JOIN message_attachment_join j ON a.ROWID = j.attachment_id
		WHERE j.message_id = ?
		ORDER BY a.ROWID
	`

	rows, err := p.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var att MessageAttachment
		var filename sql.NullString
		err := rows.Scan(&att.ID, &att.GUID, &att.MIMEType, &filename)
		if err != nil {
			continue
		}
		if filename.Valid {
			att.Filename = filename.String
		}
		attachments = append(attachments, att)
	}

	return attachments, nil
}

// createThumbnailFromMetadata creates a URLThumbnail using extracted metadata and attachments
func (p *URLProcessor) createThumbnailFromMetadata(url string, metadata *RichLinkMetadata, attachments []MessageAttachment) *URLThumbnail {
	result := &URLThumbnail{
		URL:         url,
		Title:       metadata.Title,
		Description: metadata.Summary,
		Success:     false,
	}

	// Generate thumbnail filename
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	thumbnailPath := filepath.Join(p.cacheDir, hash+".png")

	// Try 1: If we have an image attachment, try to copy it
	if metadata.HasImage && metadata.ImageIndex < len(attachments) {
		att := attachments[metadata.ImageIndex]
		if p.copyAttachmentAsImage(att, url, result) {
			return result
		}
	}

	// Try 2: If we have an image URL from metadata, download it
	if metadata.HasImage && metadata.ImageURL != "" {
		if p.downloadImageFromURL(metadata.ImageURL, thumbnailPath, result) {
			return result
		}
	}

	// Try 3: If we have an icon URL, download it
	if metadata.HasIcon && metadata.IconURL != "" {
		if p.downloadImageFromURL(metadata.IconURL, thumbnailPath, result) {
			return result
		}
	}

	// Fallback to domain card with extracted title
	if p.generateDomainCard(url, thumbnailPath, result) {
		result.ThumbnailPath = thumbnailPath
		result.Success = true
	}

	return result
}

// copyAttachmentAsImage copies an attachment file as an image thumbnail
func (p *URLProcessor) copyAttachmentAsImage(att MessageAttachment, url string, result *URLThumbnail) bool {
	// Generate thumbnail filename
	hash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	thumbnailPath := filepath.Join(p.cacheDir, hash+".png")

	// Try to find the attachment file in the attachments directory
	possiblePaths := []string{
		filepath.Join(p.config.AttachmentsPath, "Attachments", att.GUID),
		filepath.Join(p.config.AttachmentsPath, att.GUID),
	}

	for _, sourcePath := range possiblePaths {
		if _, err := os.Stat(sourcePath); err == nil {
			// Copy the file
			if p.copyAndConvertImage(sourcePath, thumbnailPath) {
				result.ThumbnailPath = thumbnailPath
				result.Success = true
				return true
			}
		}
	}

	return false
}

// copyAndConvertImage copies and converts an image to PNG format
func (p *URLProcessor) copyAndConvertImage(sourcePath, targetPath string) bool {
	// Use ImageMagick to convert and optimize
	cmd := exec.Command("magick", sourcePath, "-resize", "400x400>", "-quality", "85", targetPath)
	return cmd.Run() == nil
}

// downloadImageFromURL downloads an image from a URL and converts it to PNG
func (p *URLProcessor) downloadImageFromURL(imageURL, targetPath string, result *URLThumbnail) bool {
	fmt.Printf("📥 Downloading image from: %s\n", imageURL)

	// Create temporary file for download
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("url_image_%x", md5.Sum([]byte(imageURL))))
	defer os.Remove(tmpFile)

	// Download the image using curl
	cmd := exec.Command("curl", "-L", "-s", "--max-time", "10", "-o", tmpFile, imageURL)
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  Failed to download image: %v\n", err)
		return false
	}

	// Check if file was downloaded
	if stat, err := os.Stat(tmpFile); err != nil || stat.Size() == 0 {
		fmt.Printf("⚠️  Downloaded file is empty or missing\n")
		return false
	}

	// Convert and resize the image
	if p.copyAndConvertImage(tmpFile, targetPath) {
		result.ThumbnailPath = targetPath
		result.Success = true
		fmt.Printf("✅ Downloaded and converted image from: %s\n", imageURL)
		return true
	}

	fmt.Printf("⚠️  Failed to convert downloaded image\n")
	return false
}

// Helper functions for parsing plist output
func extractPlistValue(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func parseInt(s string) int {
	if s == "" {
		return -1
	}
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// extractAllURLs finds all HTTP/HTTPS URLs in the plist output
func extractAllURLs(text string) []string {
	re := regexp.MustCompile(`"(https://[^"]+)"`)
	matches := re.FindAllStringSubmatch(text, -1)
	var urls []string
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, match[1])
		}
	}
	return urls
}

// isPreviewImageURL determines if a URL is likely a preview image (not an icon)
func isPreviewImageURL(url string) bool {
	// High priority: CDN preview services
	if strings.Contains(url, "cdn-link-previews") {
		return true
	}
	if strings.Contains(url, "ytimg.com") && !strings.Contains(url, "favicon") {
		return true
	}

	// Medium priority: Image files that aren't obviously icons
	if strings.Contains(url, ".jpg") || strings.Contains(url, ".jpeg") ||
	   strings.Contains(url, ".png") || strings.Contains(url, ".webp") ||
	   strings.Contains(url, ".gif") {
		// Exclude small images and favicons
		if strings.Contains(url, "favicon") || strings.Contains(url, "icon") {
			return false
		}
		// Exclude obviously small dimensions
		if strings.Contains(url, "32x32") || strings.Contains(url, "16x16") ||
		   strings.Contains(url, "64x64") {
			return false
		}
		return true
	}

	return false
}

// isIconURL determines if a URL is likely an icon/favicon
func isIconURL(url string) bool {
	return strings.Contains(url, "favicon") ||
		   strings.Contains(url, "icon") ||
		   strings.Contains(url, "32x32") ||
		   strings.Contains(url, "16x16") ||
		   strings.Contains(url, "64x64")
}

// reconstructPreviewURL attempts to reconstruct preview image URLs for services that store them as attachments
func (p *URLProcessor) reconstructPreviewURL(plistOutput, originalURL string) string {
	// Apple Music/iTunes Store reconstruction
	if strings.Contains(originalURL, "music.apple.com") || strings.Contains(originalURL, "itunes.apple.com") {
		return p.reconstructAppleMusicArtwork(plistOutput, originalURL)
	}

	// Could add other services here as needed
	return ""
}

// reconstructAppleMusicArtwork builds Apple Music artwork URL from metadata
func (p *URLProcessor) reconstructAppleMusicArtwork(plistOutput, originalURL string) string {
	// For Apple Music, we can try using their public API to get artwork
	// However, playlist artwork is often not available via direct URL reconstruction
	//
	// Alternative approach: Use a generic placeholder or extract from the original URL
	// For now, return empty to let the system fall back to other methods

	// Could potentially parse the playlist ID and try:
	// https://tools.applemediaservices.com/api/artwork/[country]/music/[id]/400x400bb.jpg
	// But this requires complex parsing and isn't guaranteed to work

	return ""
}

// fetchOpenGraphThumbnail attempts to fetch Open Graph metadata and image
func (p *URLProcessor) fetchOpenGraphThumbnail(urlStr, outputPath string, result *URLThumbnail) bool {
	fmt.Printf("🔍 Fetching metadata for: %s\n", urlStr)

	// Use curl to fetch the webpage and extract Open Graph data
	metadata := p.extractWebMetadata(urlStr)
	if metadata.Title != "" {
		result.Title = metadata.Title
	} else {
		result.Title = p.extractDomainTitle(urlStr)
	}
	result.Description = metadata.Description

	// Try to download Open Graph image if available
	if metadata.ImageURL != "" {
		fmt.Printf("📸 Downloading Open Graph image: %s\n", metadata.ImageURL)
		if p.downloadImage(metadata.ImageURL, outputPath) {
			return true
		}
	}

	// Try to get favicon as fallback
	if metadata.FaviconURL != "" {
		fmt.Printf("🎭 Downloading favicon: %s\n", metadata.FaviconURL)
		if p.downloadAndResizeFavicon(metadata.FaviconURL, outputPath, result.Title, result.Description) {
			return true
		}
	}

	return false
}

// WebMetadata holds extracted webpage metadata
type WebMetadata struct {
	Title       string
	Description string
	ImageURL    string
	FaviconURL  string
}

// extractWebMetadata fetches and parses webpage metadata
func (p *URLProcessor) extractWebMetadata(urlStr string) WebMetadata {
	metadata := WebMetadata{}

	// Use curl to fetch HTML content
	cmd := exec.Command("curl", "-L", "-A", "Mozilla/5.0 (compatible; iMessages-Book)", "--max-time", "10", urlStr)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠️  Failed to fetch %s: %v\n", urlStr, err)
		return metadata
	}

	html := string(output)

	// Parse URL to get base URL for relative links
	parsedURL, _ := url.Parse(urlStr)
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Extract Open Graph title
	if ogTitle := p.extractMetaContent(html, `property=["\']og:title["\']`); ogTitle != "" {
		metadata.Title = ogTitle
	} else if title := p.extractHTMLTitle(html); title != "" {
		metadata.Title = title
	}

	// Extract Open Graph description
	if ogDesc := p.extractMetaContent(html, `property=["\']og:description["\']`); ogDesc != "" {
		metadata.Description = ogDesc
	} else if desc := p.extractMetaContent(html, `name=["\']description["\']`); desc != "" {
		metadata.Description = desc
	}

	// Extract Open Graph image
	ogImage := p.extractMetaContent(html, `property=["\']og:image["\']`)
	if ogImage != "" {
		// Convert relative URLs to absolute
		if strings.HasPrefix(ogImage, "/") {
			metadata.ImageURL = baseURL + ogImage
		} else if strings.HasPrefix(ogImage, "http") {
			metadata.ImageURL = ogImage
		} else {
			metadata.ImageURL = baseURL + "/" + ogImage
		}
	}

	if favicon := p.extractFaviconURL(html, baseURL); favicon != "" {
		metadata.FaviconURL = favicon
	} else {
		// Default favicon location
		metadata.FaviconURL = baseURL + "/favicon.ico"
	}

	return metadata
}

// extractMetaContent extracts content from meta tags using regex
func (p *URLProcessor) extractMetaContent(html, pattern string) string {
	// Create regex to find meta tag with the specified property/name
	fullPattern := fmt.Sprintf(`<meta[^>]+%s[^>]+content=["\']([^"\']+)["\']`, pattern)
	re := regexp.MustCompile(fullPattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractHTMLTitle extracts the page title
func (p *URLProcessor) extractHTMLTitle(html string) string {
	re := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// extractFaviconURL extracts favicon URL from HTML
func (p *URLProcessor) extractFaviconURL(html, baseURL string) string {
	// Look for various favicon link tags
	patterns := []string{
		`<link[^>]+rel=["\']icon["\'][^>]+href=["\']([^"\']+)["\']`,
		`<link[^>]+rel=["\']shortcut icon["\'][^>]+href=["\']([^"\']+)["\']`,
		`<link[^>]+href=["\']([^"\']+)["\'][^>]+rel=["\']icon["\']`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			faviconURL := strings.TrimSpace(matches[1])
			// Convert relative URLs to absolute
			if strings.HasPrefix(faviconURL, "/") {
				return baseURL + faviconURL
			} else if !strings.HasPrefix(faviconURL, "http") {
				return baseURL + "/" + faviconURL
			}
			return faviconURL
		}
	}
	return ""
}

// downloadImage downloads an image from URL
func (p *URLProcessor) downloadImage(imageURL, outputPath string) bool {
	cmd := exec.Command("curl", "-L", "--max-time", "15", "-o", outputPath, imageURL)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("⚠️  Failed to download image %s: %v\n", imageURL, err)
		return false
	}

	// Verify the file was created and has content
	if stat, err := os.Stat(outputPath); err != nil || stat.Size() == 0 {
		os.Remove(outputPath) // Clean up empty file
		return false
	}

	// Resize/optimize the downloaded image
	return p.optimizeDownloadedImage(outputPath)
}

// downloadAndResizeFavicon downloads a favicon and creates a card with it
func (p *URLProcessor) downloadAndResizeFavicon(faviconURL, outputPath, title, description string) bool {
	// Download favicon to temporary location
	tempFavicon := filepath.Join(p.cacheDir, "temp_favicon.ico")
	defer os.Remove(tempFavicon)

	cmd := exec.Command("curl", "-L", "--max-time", "10", "-o", tempFavicon, faviconURL)
	err := cmd.Run()
	if err != nil {
		return false
	}

	// Verify favicon was downloaded
	if stat, err := os.Stat(tempFavicon); err != nil || stat.Size() == 0 {
		return false
	}

	// Create a card with the favicon and text
	return p.createFaviconCard(tempFavicon, outputPath, title, description)
}

// optimizeDownloadedImage resizes and optimizes a downloaded image
func (p *URLProcessor) optimizeDownloadedImage(imagePath string) bool {
	// Check if the file is actually an image first
	cmd := exec.Command("file", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If it's not an image format we can handle, just return success
	outputStr := strings.ToLower(string(output))
	if !strings.Contains(outputStr, "image") && !strings.Contains(outputStr, "jpeg") && !strings.Contains(outputStr, "png") && !strings.Contains(outputStr, "gif") {
		fmt.Printf("⚠️  File %s is not a recognized image format\n", imagePath)
		return false
	}

	// Use ImageMagick to resize and optimize
	cmd = exec.Command("magick", imagePath,
		"-resize", "800x600>", // Resize maintaining aspect ratio
		"-quality", "85",
		"-strip", // Remove metadata
		"-auto-orient", // Fix orientation
		imagePath) // Overwrite original

	err = cmd.Run()
	if err != nil {
		fmt.Printf("⚠️  Failed to optimize image %s: %v\n", imagePath, err)
		// Don't return false - the image might still be usable
		return true
	}

	return true
}

// createFaviconCard creates a nice card with favicon and text
func (p *URLProcessor) createFaviconCard(faviconPath, outputPath, title, description string) bool {
	// Use ImageMagick to create a card with favicon and text
	if description == "" {
		description = "Web Link"
	}

	// Truncate long titles
	if len(title) > 40 {
		title = title[:37] + "..."
	}

	cmd := exec.Command("magick",
		"-size", "400x200",
		"xc:white",
		"(", faviconPath, "-resize", "32x32", ")",
		"-gravity", "center",
		"-geometry", "+0-40", // Position favicon above center
		"-composite",
		"-gravity", "center",
		"-pointsize", "16",
		"-fill", "black",
		"-annotate", "+0+20", title, // Title below favicon
		"-pointsize", "12",
		"-fill", "gray",
		"-annotate", "+0+40", description, // Description below title
		"-border", "1x1",
		"-bordercolor", "lightgray",
		outputPath)

	err := cmd.Run()
	if err != nil {
		fmt.Printf("⚠️  Failed to create favicon card: %v\n", err)
		return false
	}

	return true
}

// takeScreenshot captures a screenshot of the webpage
func (p *URLProcessor) takeScreenshot(urlStr, outputPath string, result *URLThumbnail) bool {
	fmt.Printf("📸 Taking screenshot of: %s\n", urlStr)

	// Use headless browser approach if available
	// For macOS, we can try using built-in screenshot tools

	// Try using playwright or similar tool if installed
	if p.tryPlaywrightScreenshot(urlStr, outputPath) {
		result.Title = p.extractDomainTitle(urlStr)
		result.Description = "Website screenshot"
		return true
	}

	// Try using WebKit2PNG if available
	if p.tryWebKit2PNG(urlStr, outputPath) {
		result.Title = p.extractDomainTitle(urlStr)
		result.Description = "Website screenshot"
		return true
	}

	return false
}

// tryPlaywrightScreenshot attempts to use Playwright for screenshots
func (p *URLProcessor) tryPlaywrightScreenshot(urlStr, outputPath string) bool {
	// Check if playwright is available
	if _, err := exec.LookPath("playwright"); err != nil {
		return false
	}

	// Create a simple Playwright script
	script := fmt.Sprintf(`
const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.setViewportSize({ width: 1200, height: 800 });

  try {
    await page.goto('%s', { waitUntil: 'networkidle', timeout: 30000 });
    await page.screenshot({ path: '%s', fullPage: false });
    console.log('Screenshot saved');
  } catch (error) {
    console.error('Screenshot failed:', error);
    process.exit(1);
  } finally {
    await browser.close();
  }
})();
`, urlStr, outputPath)

	scriptPath := filepath.Join(p.cacheDir, "screenshot.js")
	err := os.WriteFile(scriptPath, []byte(script), 0644)
	if err != nil {
		return false
	}
	defer os.Remove(scriptPath)

	cmd := exec.Command("node", scriptPath)
	cmd.Dir = p.cacheDir

	// Set timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err == nil {
			// Verify screenshot was created
			if _, err := os.Stat(outputPath); err == nil {
				return true
			}
		}
	case <-time.After(45 * time.Second):
		cmd.Process.Kill()
	}

	return false
}

// tryWebKit2PNG attempts to use webkit2png for screenshots
func (p *URLProcessor) tryWebKit2PNG(urlStr, outputPath string) bool {
	// Check if webkit2png is available
	if _, err := exec.LookPath("webkit2png"); err != nil {
		return false
	}

	tempDir := filepath.Join(p.cacheDir, "temp")
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	cmd := exec.Command("webkit2png",
		"--clipped",
		"--clipwidth=1200",
		"--clipheight=800",
		"--delay=3",
		"--dir="+tempDir,
		urlStr)

	err := cmd.Run()
	if err != nil {
		return false
	}

	// webkit2png creates files with specific naming
	parsedURL, _ := url.Parse(urlStr)
	expectedFile := filepath.Join(tempDir, parsedURL.Host+"-clipped.png")

	if _, err := os.Stat(expectedFile); err == nil {
		// Move to our desired location
		return os.Rename(expectedFile, outputPath) == nil
	}

	return false
}

// generateDomainCard creates a simple text-based card for the domain
func (p *URLProcessor) generateDomainCard(urlStr, outputPath string, result *URLThumbnail) bool {
	fmt.Printf("🎨 Generating domain card for: %s\n", urlStr)

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	domain := parsedURL.Host
	if strings.HasPrefix(domain, "www.") {
		domain = domain[4:]
	}

	result.Title = domain
	result.Description = "Web link"

	// Use ImageMagick to create a simple card
	cmd := exec.Command("magick",
		"-size", "400x200",
		"xc:white",
		"-gravity", "center",
		"-pointsize", "24",
		"-fill", "black",
		"-annotate", "+0-20", domain,
		"-pointsize", "14",
		"-fill", "gray",
		"-annotate", "+0+20", "🔗 Web Link",
		"-border", "2x2",
		"-bordercolor", "lightgray",
		outputPath)

	err = cmd.Run()
	if err != nil {
		fmt.Printf("⚠️  Failed to generate domain card: %v\n", err)
		return false
	}

	return true
}

// extractDomainTitle extracts a clean title from URL
func (p *URLProcessor) extractDomainTitle(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "Web Link"
	}

	domain := parsedURL.Host
	if strings.HasPrefix(domain, "www.") {
		domain = domain[4:]
	}

	// Capitalize first letter
	if len(domain) > 0 {
		domain = strings.ToUpper(domain[:1]) + domain[1:]
	}

	return domain
}

// ReplaceURLsWithImages replaces URLs in text with image references
func (p *URLProcessor) ReplaceURLsWithImages(text string, thumbnails map[string]*URLThumbnail) string {
	if text == "" {
		return text
	}

	result := text
	urls := p.FindURLsInText(text)

	for _, urlStr := range urls {
		if thumbnail, exists := thumbnails[urlStr]; exists && thumbnail.Success {
			// Create relative path for the markdown using forward slashes for LaTeX compatibility
			relPath := "Attachments/url-thumbnails/" + filepath.Base(thumbnail.ThumbnailPath)

			// Replace URL with image reference that works with LaTeX
			replacement := fmt.Sprintf("\\messageimage{%s}", relPath)
			result = strings.ReplaceAll(result, urlStr, replacement)
		}
	}

	return result
}