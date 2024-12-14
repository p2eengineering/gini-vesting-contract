package vesting

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

type UserVestings []string

type UserVestingsData struct {
	UserVestings []string `json:"userVestings"`
}

type Beneficiary struct {
	TotalAllocations string `json:"totalAllocations"`
	ClaimedAmount    string `json:"claimedAmount"`
}

type VestingPeriod struct {
	TotalSupply         string `json:"totalSupply"`
	CliffStartTimestamp uint64 `json:"cliffStartTimestamp"`
	StartTimestamp      uint64 `json:"startTimestamp"`
	EndTimestamp        uint64 `json:"endTimestamp"`
	Duration            uint64 `json:"duration"`
	TGE                 uint64 `json:"tge"`
}

type VestingData struct {
	VestingPeriod *VestingPeriod `json:"vestingPeriod"`
	ClaimedAmount string         `json:"claimedAmount"`
}

type ClaimsWithAllVestings struct {
	TotalAmount  string   `json:"totalAmount"`
	UserVestings []string `json:"userVestings"`
	Amounts      []string `json:"amounts"`
}

type VestingDurationsData struct {
	UserVestings     []string `json:"userVestings"`
	VestingDurations []uint64 `json:"vestingDurations"`
}

type AllocationsWithAllVestings struct {
	UserVestings     []string `json:"userVestings"`
	TotalAllocations []string `json:"totalAllocations"`
}

type TotalClaimsWithAllVestings struct {
	UserVestings []string `json:"userVestings"`
	TotalClaims  []string `json:"totalClaims"`
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

// GetTotalClaimsForAll retrieves the total claims for all vestings from the state
func GetTotalClaimsForAll(ctx kalpsdk.TransactionContextInterface) (*big.Int, error) {
	// Key for total claims for all vestings
	totalClaimsKey := "total_claims_for_all"

	// Retrieve the state
	totalClaimsAsBytes, err := ctx.GetState(totalClaimsKey)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get total claims with Key %s", totalClaimsKey), err)
	}

	totalClaims := big.NewInt(0)
	if totalClaimsAsBytes != nil {
		_, success := totalClaims.SetString(string(totalClaimsAsBytes), 10)
		if !success {
			return nil, NewCustomError(http.StatusInternalServerError, "failed to parse claimed amount for all", nil)
		}
	}

	return totalClaims, nil
}

// SetTotalClaimsForAll sets the total claims for all vestings in the state
func SetTotalClaimsForAll(ctx kalpsdk.TransactionContextInterface, totalClaims *big.Int) error {
	totalClaimsKey := "total_claims_for_all"

	totalClaimsAsBytes, err := totalClaims.MarshalText()
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to marshal total claims", err)
	}

	err = ctx.PutStateWithoutKYC(totalClaimsKey, totalClaimsAsBytes)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set total claims", err)
	}

	return nil
}

// GetTotalClaims retrieves the total claims for a specific vesting ID from the state
func GetTotalClaims(ctx kalpsdk.TransactionContextInterface, vestingID string) (*big.Int, error) {
	totalClaimsKey := fmt.Sprintf("total_claims_%s", vestingID)

	totalClaimsAsBytes, err := ctx.GetState(totalClaimsKey)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get total claims with Key %s", totalClaimsKey), err)
	}

	totalClaims := big.NewInt(0)
	if totalClaimsAsBytes != nil {
		_, success := totalClaims.SetString(string(totalClaimsAsBytes), 10)
		if !success {
			return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to parse claimed amount for vesting ID %s", vestingID), nil)
		}
	}

	return totalClaims, nil
}

// SetTotalClaims sets the total claims for a specific vesting ID in the state
func SetTotalClaims(ctx kalpsdk.TransactionContextInterface, vestingID string, totalClaims *big.Int) error {
	totalClaimsKey := fmt.Sprintf("total_claims_%s", vestingID)

	totalClaimsAsBytes, err := totalClaims.MarshalText()
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to marshal total claims", err)
	}

	err = ctx.PutStateWithoutKYC(totalClaimsKey, totalClaimsAsBytes)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to set total claims for vesting ID %s", vestingID), err)
	}

	return nil
}

// GetGiniTokenAddress retrieves the Gini Token address from the blockchain state
func GetGiniTokenAddress(ctx kalpsdk.TransactionContextInterface) (string, error) {
	giniTokenAddressBytes, err := ctx.GetState(giniTokenKey)
	if err != nil {
		return "", NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get Gini token address with Key %s", giniTokenKey), err)
	}

	return string(giniTokenAddressBytes), nil
}

// SetGiniTokenAddress sets the Gini Token address in the blockchain state
func SetGiniTokenAddress(ctx kalpsdk.TransactionContextInterface, tokenAddress string) error {
	existingAddress, err := ctx.GetState(giniTokenKey)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get Gini token address with Key %s", giniTokenKey), err)
	}
	if existingAddress != nil && string(existingAddress) != "" {
		return NewCustomError(http.StatusConflict, "Gini token address is already set", nil)
	}

	// Store the new Gini token address in the state
	err = ctx.PutStateWithoutKYC(giniTokenKey, []byte(tokenAddress))
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to set Gini token address with Key %s", giniTokenKey), err)
	}

	return nil
}