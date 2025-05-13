package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRespondNotFoundError(t *testing.T) {
	testCases := []struct {
		name           string
		err            RespondNotFoundError
		expectedString string
		expectedBase   error
	}{
		{
			name:           "default message, no base error",
			err:            RespondNotFoundError{},
			expectedString: "respond not found",
			expectedBase:   nil,
		},
		{
			name:           "custom message, no base error",
			err:            RespondNotFoundError{Message: "custom respond not found"},
			expectedString: "custom respond not found",
			expectedBase:   nil,
		},
		{
			name:           "default message, with base error",
			err:            RespondNotFoundError{BaseErr: errors.New("base error")},
			expectedString: "respond not found. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "custom message, with base error",
			err:            RespondNotFoundError{Message: "custom error", BaseErr: errors.New("base error")},
			expectedString: "custom error. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Проверка строки ошибки
			require.Equal(t, tc.expectedString, tc.err.Error())

			// Проверка базовой ошибки через Unwrap
			baseErr := tc.err.Unwrap()
			if tc.expectedBase == nil {
				require.Nil(t, baseErr)
			} else {
				require.Equal(t, tc.expectedBase.Error(), baseErr.Error())
			}

			// Проверка, что ошибка реализует интерфейс error
			var err interface{} = tc.err
			_, ok := err.(error)
			require.True(t, ok, "RespondNotFoundError should implement error interface")
		})
	}
}

func TestRespondAlreadyExistsError(t *testing.T) {
	testCases := []struct {
		name           string
		err            RespondAlreadyExistsError
		expectedString string
		expectedBase   error
	}{
		{
			name:           "default message, no base error",
			err:            RespondAlreadyExistsError{},
			expectedString: "respond already exists",
			expectedBase:   nil,
		},
		{
			name:           "custom message, no base error",
			err:            RespondAlreadyExistsError{Message: "custom respond exists"},
			expectedString: "custom respond exists",
			expectedBase:   nil,
		},
		{
			name:           "default message, with base error",
			err:            RespondAlreadyExistsError{BaseErr: errors.New("base error")},
			expectedString: "respond already exists. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "custom message, with base error",
			err:            RespondAlreadyExistsError{Message: "custom error", BaseErr: errors.New("base error")},
			expectedString: "custom error. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Проверка строки ошибки
			require.Equal(t, tc.expectedString, tc.err.Error())

			// Проверка базовой ошибки через Unwrap
			baseErr := tc.err.Unwrap()
			if tc.expectedBase == nil {
				require.Nil(t, baseErr)
			} else {
				require.Equal(t, tc.expectedBase.Error(), baseErr.Error())
			}

			// Проверка, что ошибка реализует интерфейс error
			var err interface{} = tc.err
			_, ok := err.(error)
			require.True(t, ok, "RespondAlreadyExistsError should implement error interface")
		})
	}
}

func TestRespondToOwnTicketError(t *testing.T) {
	testCases := []struct {
		name           string
		err            RespondToOwnTicketError
		expectedString string
		expectedBase   error
	}{
		{
			name:           "default message, no base error",
			err:            RespondToOwnTicketError{},
			expectedString: "respond to own Ticket is not allowed",
			expectedBase:   nil,
		},
		{
			name:           "custom message, no base error",
			err:            RespondToOwnTicketError{Message: "custom own ticket error"},
			expectedString: "custom own ticket error",
			expectedBase:   nil,
		},
		{
			name:           "default message, with base error",
			err:            RespondToOwnTicketError{BaseErr: errors.New("base error")},
			expectedString: "respond to own Ticket is not allowed. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "custom message, with base error",
			err:            RespondToOwnTicketError{Message: "custom error", BaseErr: errors.New("base error")},
			expectedString: "custom error. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Проверка строки ошибки
			require.Equal(t, tc.expectedString, tc.err.Error())

			// Проверка базовой ошибки через Unwrap
			baseErr := tc.err.Unwrap()
			if tc.expectedBase == nil {
				require.Nil(t, baseErr)
			} else {
				require.Equal(t, tc.expectedBase.Error(), baseErr.Error())
			}

			// Проверка, что ошибка реализует интерфейс error
			var err interface{} = tc.err
			_, ok := err.(error)
			require.True(t, ok, "RespondToOwnTicketError should implement error interface")
		})
	}
}
