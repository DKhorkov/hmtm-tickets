package errors

import "fmt"

type RespondNotFoundError struct {
	Message string
	BaseErr error
}

func (e RespondNotFoundError) Error() string {
	template := "respond not found"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e RespondNotFoundError) Unwrap() error {
	return e.BaseErr
}

type RespondAlreadyExistsError struct {
	Message string
	BaseErr error
}

func (e RespondAlreadyExistsError) Error() string {
	template := "respond already exists"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e RespondAlreadyExistsError) Unwrap() error {
	return e.BaseErr
}

type RespondToOwnTicketError struct {
	Message string
	BaseErr error
}

func (e RespondToOwnTicketError) Error() string {
	template := "respond to own Ticket is not allowed"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e RespondToOwnTicketError) Unwrap() error {
	return e.BaseErr
}
