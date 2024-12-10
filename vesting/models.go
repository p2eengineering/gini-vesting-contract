package vesting

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

type UserVestings []string

type Beneficiary struct {
	TotalAllocations string
	ClaimedAmount    string
}

type VestingPeriod struct {
	TotalSupply         string
	CliffStartTimestamp uint64
	StartTimestamp      uint64
	EndTimestamp        uint64
	Duration            uint64
	TGE                 uint64
}

// GetBeneficiary retrieves a Beneficiary by ID
func GetBeneficiary(ctx kalpsdk.TransactionContextInterface, beneficiaryKey string) (*Beneficiary, error) {
	beneficiaryAsBytes, err := ctx.GetState(beneficiaryKey)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get beneficiary with Key %s", beneficiaryKey), err)
	}
	if beneficiaryAsBytes == nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("beneficiary with Key %s does not exist", beneficiaryKey), nil)
	}

	var beneficiary Beneficiary
	err = json.Unmarshal(beneficiaryAsBytes, &beneficiary)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, "failed to unmarshal beneficiary", err)
	}

	return &beneficiary, nil
}

// SetBeneficiary sets a Beneficiary in the state
func SetBeneficiary(ctx kalpsdk.TransactionContextInterface, beneficiaryKey string, beneficiary *Beneficiary) error {
	beneficiaryAsBytes, err := json.Marshal(beneficiary)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to marshal beneficiaries", err)
	}

	err = ctx.PutStateWithoutKYC(beneficiaryKey, beneficiaryAsBytes)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set beneficiary", err)
	}

	return nil
}

// GetVestingPeriod retrieves a VestingPeriod by Key
func GetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingKey string) (*VestingPeriod, error) {
	vestingAsBytes, err := ctx.GetState(vestingKey)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get vesting with Key %s", vestingKey), err)
	}
	if vestingAsBytes == nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("vesting with Key %s does not exist", vestingKey), nil)
	}

	var vesting *VestingPeriod
	err = json.Unmarshal(vestingAsBytes, &vesting)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to unmarshal vesting"), err)
	}

	return vesting, nil
}

// SetVestingPeriod sets a VestingPeriod in the state
func SetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingKey string, vesting *VestingPeriod) error {
	vestingAsBytes, err := json.Marshal(vesting)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to marshal vesting", err)
	}

	err = ctx.PutStateWithoutKYC(vestingKey, vestingAsBytes)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set vesting", err)
	}

	return nil
}

func GetUserVesting(ctx kalpsdk.TransactionContextInterface, vestingKey string) (UserVestings, error) {
	userVestingJSON, err := ctx.GetState(vestingKey)
	if err != nil {
		return nil, NewCustomError(http.StatusNotFound, fmt.Sprintf("Failed to get user vestings for %s", vestingKey), err)
	}

	// If there is no vesting JSON, initialize an empty list
	if userVestingJSON == nil {
		return UserVestings{}, nil
	}

	// Unmarshal the JSON into a slice of strings
	var userVestingList UserVestings
	err = json.Unmarshal(userVestingJSON, &userVestingList)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal user vesting list for %s", vestingKey), err)
	}

	return userVestingList, nil
}

func SetUserVesting(ctx kalpsdk.TransactionContextInterface, beneficiary string, userVestingList UserVestings) error {
	updatedUserVestingJSON, err := json.Marshal(userVestingList)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal updated user vesting list for %s", beneficiary), err)
	}

	// Generate the key to store user vesting in the state
	vestingKey := fmt.Sprintf("uservestings_%s", beneficiary)

	// Store the updated vesting list on the blockchain ledger
	err = ctx.PutStateWithoutKYC(vestingKey, updatedUserVestingJSON)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to set updated user vesting list for %s", beneficiary), err)
	}

	return nil
}
