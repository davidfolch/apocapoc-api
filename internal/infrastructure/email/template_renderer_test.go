package email

import (
	"strings"
	"testing"
)

func TestNewTemplateRenderer(t *testing.T) {
	renderer := NewTemplateRenderer("Test App", "https://example.com", "support@example.com")

	if renderer == nil {
		t.Fatal("Expected renderer to be created")
	}

	if renderer.appName != "Test App" {
		t.Errorf("Expected app name 'Test App', got '%s'", renderer.appName)
	}
}

func TestTemplateRenderer_Render(t *testing.T) {
	renderer := NewTemplateRenderer("Test App", "https://example.com", "support@example.com")

	template := `Hello {{.Data.Name}}, welcome to {{.AppName}}!`
	data := map[string]interface{}{
		"Name": "John",
	}

	result, err := renderer.Render(template, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expected := "Hello John, welcome to Test App!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTemplateRenderer_RenderWithAllVariables(t *testing.T) {
	renderer := NewTemplateRenderer("My App", "https://myapp.com", "help@myapp.com")

	template := `
App: {{.AppName}}
URL: {{.AppURL}}
Support: {{.SupportEmail}}
User: {{.Data.User}}
`
	data := map[string]interface{}{
		"User": "Alice",
	}

	result, err := renderer.Render(template, data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if !strings.Contains(result, "My App") {
		t.Error("Expected result to contain app name")
	}
	if !strings.Contains(result, "https://myapp.com") {
		t.Error("Expected result to contain app URL")
	}
	if !strings.Contains(result, "help@myapp.com") {
		t.Error("Expected result to contain support email")
	}
	if !strings.Contains(result, "Alice") {
		t.Error("Expected result to contain user name")
	}
}

func TestTemplateRenderer_RenderInvalidTemplate(t *testing.T) {
	renderer := NewTemplateRenderer("Test App", "https://example.com", "support@example.com")

	template := `{{.Data.Invalid}}`
	data := map[string]interface{}{}

	result, err := renderer.Render(template, data)
	if err != nil {
		t.Fatalf("Template should render even with missing data: %v", err)
	}

	if result != "<no value>" {
		t.Logf("Got result: %s", result)
	}
}

func TestTemplateRenderer_RenderSyntaxError(t *testing.T) {
	renderer := NewTemplateRenderer("Test App", "https://example.com", "support@example.com")

	template := `{{.Data.Name`
	data := map[string]interface{}{}

	_, err := renderer.Render(template, data)
	if err == nil {
		t.Error("Expected error for invalid template syntax")
	}
}
