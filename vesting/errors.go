package vesting

import (
	"errors"
	"fmt"
)

var (
	ErrNoBeneficiaries                    = errors.New("NoBeneficiaries")
	ErrCannotBeZero                       = errors.New("CannotBeZero")
	ErrContractAddressAlreadySet          = errors.New("ContractAddressAlreadySet")
	ErrNonPositiveVestingAmount           = errors.New("NonPositiveVestingAmount")
	ErrNothingToClaim                     = errors.New("NothingToClaim")
	ErrTokenAlreadySet                    = errors.New("TokenAlreadySet")
	ErrDurationCannotBeZeroForClaimAmount = errors.New("DurationCannotBeZero")
	ErrTotalAllocationCannotBeNonPositive = errors.New("TotalAllocationCannotBeNonPositive")
	ErrInitialUnlockCannotBeNegative      = errors.New("InitialUnlockCannotBeNegative")
	ErrEmptyAddress                       = errors.New("ErrEmptyAddress")
)

type CustomError struct {
	Code    int
	Message string
	Err     error
}

func ErrStartTimestampLessThanCurrentTimeStamp(startTimeStamp, currentTime uint64) error {
	return fmt.Errorf("start timestamp %d is less than the current time %d", startTimeStamp, currentTime)
}

func ErrRegexValidationFailed(field string, address string, err error) error {
	return fmt.Errorf("failed to validate %s for address %s due to regex error: %v", field, address, err)
}

func ErrInvalidUserAddress(userAddress string) error {
	return fmt.Errorf("InvalidUserAddress for userAddress %s", userAddress)
}

func ErrDurationCannotBeZero(vestingID string) error {
	return fmt.Errorf("DurationCannotBeZero for vestingID %s", vestingID)
}

func ErrTotalSupplyCannotBeNonPositive(vestingID string) error {
	return fmt.Errorf("TotalSupplyCannotBeNonPositive for vestingID %s", vestingID)
}

func ErrTotalSupplyCannotBeNegative(vestingID string) error {
	return fmt.Errorf("TotalSupplyCannotBeNegative for vestingID %s", vestingID)
}

func ErrInvalidAmount(entity, value, amount string) error {
	return fmt.Errorf("InvalidAmount for %s with value %s for amount %s", entity, value, amount)
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
	return fmt.Errorf("BeneficiaryAlreadyExists for beneficiary: %s", beneficiary)
}

func ErrUserVestingsAlreadyExists(beneficiary string) error {
	return fmt.Errorf("UserVestingsAlreadyExists for beneficiary: %s", beneficiary)
}

func ErrClaimAmountExceedsVestingAmount(vestingID, beneficiaryAddress, claimAmount, beneficiaryTotalAllocations string) error {
	return fmt.Errorf("ClaimAmountExceedsVestingAmount for vesting ID %s and beneficiary %s: claimAmount=%s, totalAllocations=%s",
		vestingID, beneficiaryAddress, claimAmount, beneficiaryTotalAllocations)
}

func ErrInvalidVestingID(vestingID string) error {
	return fmt.Errorf("InvalidVestingID for vestingID: %s", vestingID)
}

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func NewCustomError(code int, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
