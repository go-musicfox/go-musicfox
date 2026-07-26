package model

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/style"
)

// FormField defines a single input field in a form.
type FormField struct {
	Key         string
	Label       string
	Placeholder string
	Required    bool
	Validate    func(string) error
}

// Form is a composable multi-field input widget with validation and focus management.
type Form struct {
	fields []FormField
	inputs []textinput.Model
	errors []error

	focusedIdx int
	focused    bool
	submitted  bool

	width  int
	height int
}

// NewForm creates a new Form with the given field definitions.
// Returns a Form ready for embedding in a host model.
func NewForm(fields []FormField) *Form {
	inputs := make([]textinput.Model, len(fields))
	errors := make([]error, len(fields))

	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		ti.CharLimit = 256
		inputs[i] = ti
	}

	return &Form{
		fields:     fields,
		inputs:     inputs,
		errors:     errors,
		focusedIdx: 0,
		focused:    false,
		submitted:  false,
	}
}

// Focus marks the form as focused and focuses the first field.
func (f *Form) Focus() {
	f.focused = true
	if len(f.inputs) > 0 {
		f.inputs[f.focusedIdx].Focus()
	}
}

// Blur marks the form as blurred and blurs all fields.
func (f *Form) Blur() {
	f.focused = false
	for i := range f.inputs {
		f.inputs[i].Blur()
	}
}

// Focused returns whether the form is currently focused.
func (f *Form) Focused() bool {
	return f.focused
}

// SetSize sets the form's display dimensions.
func (f *Form) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// Update handles input events for the form.
func (f *Form) Update(msg tea.Msg) tea.Cmd {
	if !f.focused || len(f.inputs) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "tab", "down":
			// Move to next field
			return f.nextField()
		case "shift+tab", "up":
			// Move to previous field
			return f.prevField()
		case "enter":
			// Submit if on last field or all fields are valid
			return f.trySubmit()
		case "esc":
			// Cancel - host should handle this
			return nil
		}
	}

	// Forward message to focused input
	var cmd tea.Cmd
	f.inputs[f.focusedIdx], cmd = f.inputs[f.focusedIdx].Update(msg)
	return cmd
}

// nextField moves focus to the next field.
func (f *Form) nextField() tea.Cmd {
	// Validate current field on blur
	f.validateField(f.focusedIdx)

	// Move focus
	f.inputs[f.focusedIdx].Blur()
	f.focusedIdx++
	if f.focusedIdx >= len(f.inputs) {
		f.focusedIdx = 0
	}
	return f.inputs[f.focusedIdx].Focus()
}

// prevField moves focus to the previous field.
func (f *Form) prevField() tea.Cmd {
	// Validate current field on blur
	f.validateField(f.focusedIdx)

	// Move focus
	f.inputs[f.focusedIdx].Blur()
	f.focusedIdx--
	if f.focusedIdx < 0 {
		f.focusedIdx = len(f.inputs) - 1
	}
	return f.inputs[f.focusedIdx].Focus()
}

// trySubmit attempts to submit the form if all validations pass.
func (f *Form) trySubmit() tea.Cmd {
	// Validate all fields
	for i := range f.fields {
		f.validateField(i)
	}

	// Check if form is valid
	if f.IsValid() {
		f.submitted = true
		return nil
	}

	return nil
}

// validateField validates a single field and stores the error.
func (f *Form) validateField(idx int) {
	if idx < 0 || idx >= len(f.fields) {
		return
	}

	field := f.fields[idx]
	value := f.inputs[idx].Value()
	empty := strings.TrimSpace(value) == ""

	// Check required
	if field.Required && empty {
		f.errors[idx] = ErrFieldRequired
		return
	}

	// Skip custom validation for empty optional fields: an empty value is
	// acceptable when the field is not required.
	if empty {
		f.errors[idx] = nil
		return
	}

	// Run custom validation if provided
	if field.Validate != nil {
		f.errors[idx] = field.Validate(value)
		return
	}

	// Clear error if valid
	f.errors[idx] = nil
}

// ErrFieldRequired is returned when a required field is empty.
// Its Error() message is drawn from the active locale via the default catalog.
var ErrFieldRequired = &requiredFieldError{}

// requiredFieldError is the concrete type of ErrFieldRequired. Using a
// dedicated type lets consumers identity-compare against ErrFieldRequired
// while the rendered message is localized at call time.
type requiredFieldError struct{}

func (e *requiredFieldError) Error() string {
	return T(MsgFieldRequired)
}

// IsValid returns true if all field validations pass.
func (f *Form) IsValid() bool {
	for _, err := range f.errors {
		if err != nil {
			return false
		}
	}
	return true
}

// Values returns a map of field keys to their current values.
func (f *Form) Values() map[string]string {
	values := make(map[string]string, len(f.fields))
	for i, field := range f.fields {
		values[field.Key] = f.inputs[i].Value()
	}
	return values
}

// Submitted returns true if the form has been successfully submitted.
func (f *Form) Submitted() bool {
	return f.submitted
}

// Reset clears the form submission state and all field values and errors.
func (f *Form) Reset() {
	f.submitted = false
	for i := range f.inputs {
		f.inputs[i].SetValue("")
		f.errors[i] = nil
	}
}

// View renders the form.
func (f *Form) View() string {
	if len(f.fields) == 0 {
		return ""
	}

	styles := style.CurrentStyleSet()

	// Calculate label column width for alignment
	labelWidth := 0
	for _, field := range f.fields {
		w := lipgloss.Width(field.Label)
		if w > labelWidth {
			labelWidth = w
		}
	}

	var b strings.Builder

	for i, field := range f.fields {
		// Label - right-aligned in fixed-width column
		labelStyle := styles.MenuItem
		if i == f.focusedIdx && f.focused {
			labelStyle = styles.SelectedItem
		}
		label := field.Label
		if field.Required {
			label = label + " *"
		}
		paddedLabel := lipgloss.NewStyle().
			Width(labelWidth + 2).
			Align(lipgloss.Right).
			Render(label)

		// Input
		inputStyle := styles.Muted
		if i == f.focusedIdx && f.focused {
			inputStyle = styles.Prompt
		}

		// Style the textinput
		tiStyles := textinput.DefaultStyles(style.HasDarkBackground())
		if i == f.focusedIdx && f.focused {
			tiStyles.Focused.Prompt = inputStyle
			tiStyles.Focused.Text = styles.Normal
			tiStyles.Focused.Placeholder = styles.Muted
		} else {
			tiStyles.Blurred.Prompt = styles.Muted
			tiStyles.Blurred.Text = styles.Muted
			tiStyles.Blurred.Placeholder = styles.Muted
		}
		f.inputs[i].SetStyles(tiStyles)

		// Set input width
		availWidth := f.width - labelWidth - 4
		if availWidth < 20 {
			availWidth = 20
		}
		f.inputs[i].SetWidth(availWidth)

		inputView := f.inputs[i].View()

		// Render line
		b.WriteString(labelStyle.Render(paddedLabel))
		b.WriteString(styles.AppBackground.Render(" "))
		b.WriteString(inputView)
		b.WriteString("\n")

		// Error message
		if f.errors[i] != nil {
			errorMsg := f.errors[i].Error()
			errorStyle := styles.Error
			padding := styles.AppBackground.Render(strings.Repeat(" ", labelWidth+3))
			b.WriteString(padding)
			b.WriteString(errorStyle.Render(errorMsg))
			b.WriteString("\n")
		}
	}

	return strings.TrimSuffix(b.String(), "\n")
}
