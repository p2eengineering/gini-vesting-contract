package vesting

import (
	"errors"
	"fmt"
)

var (
	ErrNoBeneficiaries           = errors.New("no beneficiaries provided")
	ErrArraysLengthMismatch      = errors.New("beneficiaries and amounts arrays length mismatch")
	ErrTotalSupplyReached        = errors.New("total supply reached for vesting type")
	ErrZeroVestingAmount         = errors.New("vesting amount cannot be zero")
	ErrBeneficiaryAlreadyExists  = errors.New("beneficiary already exists")
	ErrCannotBeZero              = errors.New("startTimestamp cannot be zero")
	ErrInvalidUserAddress        = errors.New("beneficiary address cannot be zero")
	ErrInvalidContractAddress    = errors.New("contract address is invalid")
	ErrContractAddressAlreadySet = errors.New("contract address is already set")
	ErrNonPositiveVestingAmount  = errors.New("vesting amount cannot be less than or equal to zero")
)

type CustomError struct {
	Code    int
	Message string
	Err     error
}

func InvalidAmountError(entity, value, amount string) error {
	return fmt.Errorf("invalid amount format for %s with value %s", entity, value)
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewCustomError(code int, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
