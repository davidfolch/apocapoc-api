package email

import (
	"bytes"
	"fmt"
	"html/template"
)

type TemplateData struct {
	AppName      string
	AppURL       string
	SupportEmail string
	Data         map[string]interface{}
}

type TemplateRenderer struct {
	appName      string
	appURL       string
	supportEmail string
}

func NewTemplateRenderer(appName, appURL, supportEmail string) *TemplateRenderer {
	return &TemplateRenderer{
		appName:      appName,
		appURL:       appURL,
		supportEmail: supportEmail,
	}
}

func (r *TemplateRenderer) Render(templateContent string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	templateData := TemplateData{
		AppName:      r.appName,
		AppURL:       r.appURL,
		SupportEmail: r.supportEmail,
		Data:         data,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
