package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTicketNotFoundError(t *testing.T) {
	testCases := []struct {
		name           string
		err            TicketNotFoundError
		expectedString string
		expectedBase   error
	}{
		{
			name:           "default message, no base error",
			err:            TicketNotFoundError{},
			expectedString: "ticket not found",
			expectedBase:   nil,
		},
		{
			name:           "custom message, no base error",
			err:            TicketNotFoundError{Message: "custom ticket not found"},
			expectedString: "custom ticket not found",
			expectedBase:   nil,
		},
		{
			name:           "default message, with base error",
			err:            TicketNotFoundError{BaseErr: errors.New("base error")},
			expectedString: "ticket not found. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "custom message, with base error",
			err:            TicketNotFoundError{Message: "custom error", BaseErr: errors.New("base error")},
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
			require.True(t, ok, "TicketNotFoundError should implement error interface")
		})
	}
}

func TestTicketAlreadyExistsError(t *testing.T) {
	testCases := []struct {
		name           string
		err            TicketAlreadyExistsError
		expectedString string
		expectedBase   error
	}{
		{
			name:           "default message, no base error",
			err:            TicketAlreadyExistsError{},
			expectedString: "ticket already exists",
			expectedBase:   nil,
		},
		{
			name:           "custom message, no base error",
			err:            TicketAlreadyExistsError{Message: "custom ticket exists"},
			expectedString: "custom ticket exists",
			expectedBase:   nil,
		},
		{
			name:           "default message, with base error",
			err:            TicketAlreadyExistsError{BaseErr: errors.New("base error")},
			expectedString: "ticket already exists. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "custom message, with base error",
			err:            TicketAlreadyExistsError{Message: "custom error", BaseErr: errors.New("base error")},
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
			require.True(t, ok, "TicketAlreadyExistsError should implement error interface")
		})
	}
}
