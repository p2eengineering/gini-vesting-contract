package vesting

import (
	"errors"
	"fmt"
)

var (
	ErrNoBeneficiaries           = errors.New("no beneficiaries provided")
	ErrCannotBeZero              = errors.New("startTimestamp cannot be zero")
	ErrInvalidUserAddress        = errors.New("beneficiary address cannot be zero")
	ErrContractAddressAlreadySet = errors.New("contract address is already set")
	ErrNonPositiveVestingAmount  = errors.New("vesting amount cannot be less than or equal to zero")
	ErrNothingToClaim            = errors.New("Nothing to claim")
	ErrTokenAlreadySet           = errors.New("Token already set")
)

type CustomError struct {
	Code    int
	Message string
	Err     error
}

func ErrInvalidAmount(entity, value, amount string) error {
	return fmt.Errorf("invalid amount format for %s with value %s", entity, value)
}

func ErrOnlyAfterVestingStart(vestingID string) error {
	return fmt.Errorf("vesting has not started yet for vesting ID %s", vestingID)
}

func ErrInvalidContractAddress(contractAddress string) error {
	return fmt.Errorf("contract address is invalid for address %s", contractAddress)
}

func ErrArraysLengthMismatch(length1, length2 int) error {
	return fmt.Errorf("ArraysLengthMismatch: length1: %d, length2: %d", length1, length2)
}

func ErrTotalSupplyReached(vestingID string) error {
	return fmt.Errorf("total supply reached for vesting type: %s", vestingID)
}

func ZeroVestingAmount(beneficiary string) error {
	return fmt.Errorf("vesting amount cannot be zero for beneficiary: %s", beneficiary)
}

func ErrBeneficiaryAlreadyExists(beneficiary string) error {
	return fmt.Errorf("beneficiary already exists: %s", beneficiary)
}

func ErrClaimAmountExceedsVestingAmount(vestingID, beneficiaryAddress, claimAmount, beneficiaryTotalAllocations string) error {
	return fmt.Errorf("claim amount exceeds vesting amount for vesting ID %s and beneficiary %s: claimAmount=%d, totalAllocations=%d",
		vestingID, beneficiaryAddress, claimAmount, beneficiaryTotalAllocations)
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
