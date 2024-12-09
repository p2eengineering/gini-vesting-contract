package vesting

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

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
func GetBeneficiary(ctx kalpsdk.TransactionContextInterface, beneficiaryID string) (*Beneficiary, error) {
	beneficiaryAsBytes, err := ctx.GetState(beneficiaryID)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get beneficiary with ID %s", beneficiaryID), err)
	}
	if beneficiaryAsBytes == nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("beneficiary with ID %s does not exist", beneficiaryID), nil)
	}

	var beneficiary Beneficiary
	err = json.Unmarshal(beneficiaryAsBytes, &beneficiary)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, "failed to unmarshal beneficiary", err)
	}

	return &beneficiary, nil
}

// SetBeneficiary sets a Beneficiary in the state
func SetBeneficiary(ctx kalpsdk.TransactionContextInterface, beneficiaryID string, beneficiary *Beneficiary) error {
	beneficiaryAsBytes, err := json.Marshal(beneficiary)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to marshal beneficiaries", err)
	}

	err = ctx.PutStateWithoutKYC(beneficiaryID, beneficiaryAsBytes)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set beneficiary", err)
	}

	return nil
}

// GetVestingPeriod retrieves a VestingPeriod by ID
func GetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingID string) (*VestingPeriod, error) {
	vestingAsBytes, err := ctx.GetState(vestingID)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get vesting with ID %s", vestingID), err)
	}
	if vestingAsBytes == nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("vesting with ID %s does not exist", vestingID), nil)
	}

	var vesting *VestingPeriod
	err = json.Unmarshal(vestingAsBytes, &vesting)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to unmarshal vesting"), err)
	}

	return vesting, nil
}

// SetVestingPeriod sets a VestingPeriod in the state
func SetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingID string, vesting *VestingPeriod) error {
	vestingAsBytes, err := json.Marshal(vesting)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to marshal vesting", err)
	}

	err = ctx.PutStateWithoutKYC(vestingID, vestingAsBytes)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set vesting", err)
	}

	return nil
}

func GetUserVesting(ctx kalpsdk.TransactionContextInterface, vestingKey string) ([]string, error) {
	userVestingJSON, err := ctx.GetState(vestingKey)
	if err != nil {
		return nil, NewCustomError(http.StatusNotFound, fmt.Sprintf("Failed to get user vestings for %s", vestingKey), err)
	}

	// If there is no vesting JSON, initialize an empty list
	if userVestingJSON == nil {
		return []string{}, nil
	}

	// Unmarshal the JSON into a slice of strings
	var userVestingList []string
	err = json.Unmarshal(userVestingJSON, &userVestingList)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal user vesting list for %s", vestingKey), err)
	}

	return userVestingList, nil
}

func SetUserVesting(ctx kalpsdk.TransactionContextInterface, beneficiary string, userVestingList []string) error {
	updatedUserVestingJSON, err := json.Marshal(userVestingList)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal updated user vesting list for %s", beneficiary), err)
	}

	// Generate the key to store user vesting in the state
	vestingKey := fmt.Sprintf("uservesting_%s", beneficiary)

	// Store the updated vesting list on the blockchain ledger
	err = ctx.PutStateWithoutKYC(vestingKey, updatedUserVestingJSON)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to set updated user vesting list for %s", beneficiary), err)
	}

	return nil
}