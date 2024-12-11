package vesting

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

type UserVestings []string

type UserVestingsData struct {
	UserVestings []string
}

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

type VestingData struct {
	VestingPeriod *VestingPeriod
	ClaimedAmount string
}

type ClaimsWithAllVestings struct {
	TotalAmount  string
	UserVestings []string
	Amounts      []string
}

type VestingDurationsData struct {
	UserVestings     []string
	VestingDurations []uint64
}

type AllocationsWithAllVestings struct {
	UserVestings     []string
	TotalAllocations []string
}

type TotalClaimsWithAllVestings struct {
	UserVestings []string
	TotalClaims  []string
}

// GetBeneficiary retrieves a Beneficiary by ID
func GetBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiaryID string) (*Beneficiary, error) {
	beneficiaryKey := fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiaryID)
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
func SetBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiaryID string, beneficiary *Beneficiary) error {
	beneficiaryKey := fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiaryID)
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
func GetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingID string) (*VestingPeriod, error) {
	vestingKey := fmt.Sprintf("vestingperiod_%s", vestingID)
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
func SetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingID string, vesting *VestingPeriod) error {
	vestingKey := fmt.Sprintf("vestingperiod_%s", vestingID)
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

func GetUserVesting(ctx kalpsdk.TransactionContextInterface, beneficiaryID string) (UserVestings, error) {
	userVestingKey := fmt.Sprintf("uservesting_%s", beneficiaryID)
	userVestingJSON, err := ctx.GetState(userVestingKey)
	if err != nil {
		return nil, NewCustomError(http.StatusNotFound, fmt.Sprintf("Failed to get user vestings for %s", userVestingKey), err)
	}

	// If there is no vesting JSON, initialize an empty list
	if userVestingJSON == nil {
		return UserVestings{}, nil
	}

	// Unmarshal the JSON into a slice of strings
	var userVestingList UserVestings
	err = json.Unmarshal(userVestingJSON, &userVestingList)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to unmarshal user vesting list for %s", userVestingKey), err)
	}

	return userVestingList, nil
}

func SetUserVesting(ctx kalpsdk.TransactionContextInterface, beneficiaryID string, userVestingList UserVestings) error {
	updatedUserVestingJSON, err := json.Marshal(userVestingList)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal updated user vesting list for %s", beneficiaryID), err)
	}

	// Generate the key to store user vesting in the state
	userVestingKey := fmt.Sprintf("uservesting_%s", beneficiaryID)

	// Store the updated vesting list on the blockchain ledger
	err = ctx.PutStateWithoutKYC(userVestingKey, updatedUserVestingJSON)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to set updated user vesting list for %s", beneficiaryID), err)
	}

	return nil
}
