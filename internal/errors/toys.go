package errors

import "fmt"

type CategoryNotFoundError struct {
	Message string
	BaseErr error
}

func (e CategoryNotFoundError) Error() string {
	template := "category with ID=%s not found"
	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}

func (e CategoryNotFoundError) Unwrap() error {
	return e.BaseErr
}

type TagNotFoundError struct {
	Message string
	BaseErr error
}

func (e TagNotFoundError) Error() string {
	template := "tag with ID=%s not found"
	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.Message, e.BaseErr)
	}

	return fmt.Sprintf(template, e.Message)
}

func (e TagNotFoundError) Unwrap() error {
	return e.BaseErr
}
