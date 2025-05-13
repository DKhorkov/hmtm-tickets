package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCategoryNotFoundError(t *testing.T) {
	testCases := []struct {
		name           string
		err            CategoryNotFoundError
		expectedString string
		expectedBase   error
	}{
		{
			name:           "default message, no base error",
			err:            CategoryNotFoundError{Message: "1"},
			expectedString: "category with ID=1 not found",
			expectedBase:   nil,
		},
		{
			name:           "custom message, no base error",
			err:            CategoryNotFoundError{Message: "42"},
			expectedString: "category with ID=42 not found",
			expectedBase:   nil,
		},
		{
			name:           "default message, with base error",
			err:            CategoryNotFoundError{Message: "1", BaseErr: errors.New("base error")},
			expectedString: "category with ID=1 not found. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "custom message, with base error",
			err:            CategoryNotFoundError{Message: "42", BaseErr: errors.New("base error")},
			expectedString: "category with ID=42 not found. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "empty message, no base error",
			err:            CategoryNotFoundError{},
			expectedString: "category with ID= not found",
			expectedBase:   nil,
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
			require.True(t, ok, "CategoryNotFoundError should implement error interface")
		})
	}
}

func TestTagNotFoundError(t *testing.T) {
	testCases := []struct {
		name           string
		err            TagNotFoundError
		expectedString string
		expectedBase   error
	}{
		{
			name:           "default message, no base error",
			err:            TagNotFoundError{Message: "1"},
			expectedString: "tag with ID=1 not found",
			expectedBase:   nil,
		},
		{
			name:           "custom message, no base error",
			err:            TagNotFoundError{Message: "42"},
			expectedString: "tag with ID=42 not found",
			expectedBase:   nil,
		},
		{
			name:           "default message, with base error",
			err:            TagNotFoundError{Message: "1", BaseErr: errors.New("base error")},
			expectedString: "tag with ID=1 not found. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "custom message, with base error",
			err:            TagNotFoundError{Message: "42", BaseErr: errors.New("base error")},
			expectedString: "tag with ID=42 not found. Base error: base error",
			expectedBase:   errors.New("base error"),
		},
		{
			name:           "empty message, no base error",
			err:            TagNotFoundError{},
			expectedString: "tag with ID= not found",
			expectedBase:   nil,
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
			require.True(t, ok, "TagNotFoundError should implement error interface")
		})
	}
}
