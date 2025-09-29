package plugins

import (
	"threadbound/internal/output"
	"threadbound/internal/plugins/html"
	"threadbound/internal/plugins/pdf"
)

// RegisterBuiltinPlugins registers all built-in output format plugins
func RegisterBuiltinPlugins() error {
	// Register PDF plugin
	pdfPlugin := pdf.NewPDFPlugin()
	if err := output.Register(pdfPlugin); err != nil {
		return err
	}

	// Register HTML plugin
	htmlPlugin := html.NewHTMLPlugin()
	if err := output.Register(htmlPlugin); err != nil {
		return err
	}

	return nil
}

// init automatically registers built-in plugins when the package is imported
func init() {
	// Register built-in plugins
	if err := RegisterBuiltinPlugins(); err != nil {
		panic("Failed to register built-in plugins: " + err.Error())
	}
}