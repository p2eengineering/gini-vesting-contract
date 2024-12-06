package vesting

import "errors"

var (
	ErrNoBeneficiaries          = errors.New("no beneficiaries provided")
	ErrArraysLengthMismatch     = errors.New("beneficiaries and amounts arrays length mismatch")
	ErrTotalSupplyReached       = errors.New("total supply reached for vesting type")
	ErrZeroVestingAmount        = errors.New("vesting amount cannot be zero")
	ErrBeneficiaryAlreadyExists = errors.New("beneficiary already exists")
)
