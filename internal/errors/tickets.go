package errors

import "fmt"

type TicketNotFoundError struct {
	Message string
	BaseErr error
}

func (e TicketNotFoundError) Error() string {
	template := "ticket not found"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e TicketNotFoundError) Unwrap() error {
	return e.BaseErr
}

type TicketAlreadyExistsError struct {
	Message string
	BaseErr error
}

func (e TicketAlreadyExistsError) Error() string {
	template := "ticket already exists"
	if e.Message != "" {
		template = e.Message
	}

	if e.BaseErr != nil {
		return fmt.Sprintf(template+". Base error: %v", e.BaseErr)
	}

	return template
}

func (e TicketAlreadyExistsError) Unwrap() error {
	return e.BaseErr
}
