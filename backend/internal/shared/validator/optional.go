package validator

// RequireOptionalContentDescription validates an optional content description field.
func RequireOptionalContentDescription(v *string, errIfInvalid error) error {
	if v != nil && (*v == "" || !IsValidContentDescription(*v)) {
		return errIfInvalid
	}
	return nil
}

// RequireOptionalContentBody validates an optional content body field.
func RequireOptionalContentBody(v *string, errIfInvalid error) error {
	if v != nil && (*v == "" || !IsValidContentBody(*v)) {
		return errIfInvalid
	}
	return nil
}

// RequireOptionalSeoTitle validates an optional SEO title field.
func RequireOptionalSeoTitle(v *string, errIfInvalid error) error {
	if v != nil && (*v == "" || !IsValidSeoTitle(*v)) {
		return errIfInvalid
	}
	return nil
}

// RequireOptionalSeoDescription validates an optional SEO description field.
func RequireOptionalSeoDescription(v *string, errIfInvalid error) error {
	if v != nil && (*v == "" || !IsValidSeoDescription(*v)) {
		return errIfInvalid
	}
	return nil
}

// RequireOptionalHTTPSURL validates an optional HTTPS URL field.
func RequireOptionalHTTPSURL(v *string, errIfInvalid error) error {
	if v != nil && (*v == "" || !IsValidHTTPSURL(*v)) {
		return errIfInvalid
	}
	return nil
}

// RequireOptionalPositiveInt validates an optional positive-int field.
func RequireOptionalPositiveInt(v *int, errIfInvalid error) error {
	if v != nil && *v <= 0 {
		return errIfInvalid
	}
	return nil
}
