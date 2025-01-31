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

func GetBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiaryID string) (*Beneficiary, error) {
	beneficiaryKey, err := ctx.CreateCompositeKey(BeneficiariesPrefix, []string{vestingID, beneficiaryID})
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for GetBeneficiary with vestingID %s and beneficiaryID with address %s", vestingID, beneficiaryID), err)
	}

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

func SetBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiaryID string, beneficiary *Beneficiary) error {
	beneficiaryKey, err := ctx.CreateCompositeKey(BeneficiariesPrefix, []string{vestingID, beneficiaryID})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for SetBeneficiary with vestingID %s and beneficiaryID with address %s", vestingID, beneficiaryID), err)
	}

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

func GetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingID string) (*VestingPeriod, error) {
	vestingKey, err := ctx.CreateCompositeKey(VestingPeriodPrefix, []string{vestingID})
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for GetVestingPeriod with vestingID %s", vestingID), err)
	}

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

func SetVestingPeriod(ctx kalpsdk.TransactionContextInterface, vestingID string, vesting *VestingPeriod) error {
	vestingKey, err := ctx.CreateCompositeKey(VestingPeriodPrefix, []string{vestingID})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for SetVestingPeriod with vestingID %s", vestingID), err)
	}

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
	userVestingKey, err := ctx.CreateCompositeKey(UserVestingsPrefix, []string{beneficiaryID})
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for GetUserVesting with beneficiaryID %s", beneficiaryID), err)
	}

	userVestingJSON, err := ctx.GetState(userVestingKey)
	if err != nil {
		return nil, NewCustomError(http.StatusNotFound, fmt.Sprintf("Failed to get user vestings for %s", userVestingKey), err)
	}

	if userVestingJSON == nil {
		return UserVestings{}, nil
	}

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

	userVestingKey, err := ctx.CreateCompositeKey(UserVestingsPrefix, []string{beneficiaryID})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for SetUserVesting with beneficiaryID %s", beneficiaryID), err)
	}

	err = ctx.PutStateWithoutKYC(userVestingKey, updatedUserVestingJSON)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("Failed to set updated user vesting list for %s", beneficiaryID), err)
	}

	return nil
}

func GetTotalClaimsForAll(ctx kalpsdk.TransactionContextInterface) (*big.Int, error) {
	totalClaimsAsBytes, err := ctx.GetState(TotalClaimsForAllKey)
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get total claims with Key %s", TotalClaimsForAllKey), err)
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

func SetTotalClaimsForAll(ctx kalpsdk.TransactionContextInterface, totalClaims *big.Int) error {
	totalClaimsAsBytes, err := totalClaims.MarshalText()
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to marshal total claims", err)
	}

	err = ctx.PutStateWithoutKYC(TotalClaimsForAllKey, totalClaimsAsBytes)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set total claims", err)
	}

	return nil
}

func GetTotalClaims(ctx kalpsdk.TransactionContextInterface, vestingID string) (*big.Int, error) {
	totalClaimsKey, err := ctx.CreateCompositeKey(TotalClaimsPrefix, []string{vestingID})
	if err != nil {
		return nil, NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for GetTotalClaims with vestingID %s", vestingID), err)
	}

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

func SetTotalClaims(ctx kalpsdk.TransactionContextInterface, vestingID string, totalClaims *big.Int) error {
	totalClaimsKey, err := ctx.CreateCompositeKey(TotalClaimsPrefix, []string{vestingID})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for SetTotalClaims with vestingID %s", vestingID), err)
	}

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

func GetGiniTokenAddress(ctx kalpsdk.TransactionContextInterface) (string, error) {
	giniTokenAddressBytes, err := ctx.GetState(giniTokenKey)
	if err != nil {
		return "", NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get Gini token address with Key %s", giniTokenKey), err)
	}

	return string(giniTokenAddressBytes), nil
}

func SetGiniTokenAddress(ctx kalpsdk.TransactionContextInterface, tokenAddress string) error {
	existingAddress, err := ctx.GetState(giniTokenKey)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to get Gini token address with Key %s", giniTokenKey), err)
	}
	if existingAddress != nil && string(existingAddress) != "" {
		return NewCustomError(http.StatusConflict, "Gini token address is already set", nil)
	}

	err = ctx.PutStateWithoutKYC(giniTokenKey, []byte(tokenAddress))
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to set Gini token address with Key %s", giniTokenKey), err)
	}

	return nil
}
