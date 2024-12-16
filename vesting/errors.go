package vesting

import (
	"errors"
	"fmt"
)

var (
	ErrNoBeneficiaries           = errors.New("NoBeneficiaries")
	ErrCannotBeZero              = errors.New("CannotBeZero")
	ErrInvalidUserAddress        = errors.New("InvalidUserAddress")
	ErrContractAddressAlreadySet = errors.New("ContractAddressAlreadySet")
	ErrNonPositiveVestingAmount  = errors.New("NonPositiveVestingAmount")
	ErrNothingToClaim            = errors.New("NothingToClaim")
	ErrTokenAlreadySet           = errors.New("TokenAlreadySet")
)

type CustomError struct {
	Code    int
	Message string
	Err     error
}

func ErrInvalidAmount(entity, value, amount string) error {
	return fmt.Errorf("InvalidAmount for %s with value %s", entity, value)
}

func ErrOnlyAfterVestingStart(vestingID string) error {
	return fmt.Errorf("OnlyAfterVestingStart for vesting ID %s", vestingID)
}

func ErrInvalidContractAddress(contractAddress string) error {
	return fmt.Errorf("InvalidContractAddress for address %s", contractAddress)
}

func ErrArraysLengthMismatch(length1, length2 int) error {
	return fmt.Errorf("ArraysLengthMismatch: length1: %d, length2: %d", length1, length2)
}

func ErrTotalSupplyReached(vestingID string) error {
	return fmt.Errorf("TotalSupplyReached for vesting type: %s", vestingID)
}

func ErrZeroVestingAmount(beneficiary string) error {
	return fmt.Errorf("ZeroVestingAmount for beneficiary: %s", beneficiary)
}

func ErrBeneficiaryAlreadyExists(beneficiary string) error {
	return fmt.Errorf("BeneficiaryAlreadyExists: %s", beneficiary)
}

func ErrUserVestingsAlreadyExists(beneficiary string) error {
	return fmt.Errorf("UserVestingsAlreadyExists: %s", beneficiary)
}

func ErrClaimAmountExceedsVestingAmount(vestingID, beneficiaryAddress, claimAmount, beneficiaryTotalAllocations string) error {
	return fmt.Errorf("ClaimAmountExceedsVestingAmount for vesting ID %s and beneficiary %s: claimAmount=%d, totalAllocations=%d",
		vestingID, beneficiaryAddress, claimAmount, beneficiaryTotalAllocations)
}

func ErrInvalidVestingID(vestingID string) error {
	return fmt.Errorf("InvalidVestingID for vestingID: %s", vestingID)
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
