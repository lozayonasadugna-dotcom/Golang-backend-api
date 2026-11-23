package validator

// Validator accumulates field-level validation errors.
type Validator struct {
	Errors map[string]string
}

// New returns an initialized validator.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid reports whether validation passed.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError records the first error for a field.
func (v *Validator) AddError(field, message string) {
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

// Check records message for field when ok is false.
func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}
