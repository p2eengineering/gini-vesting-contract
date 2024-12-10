package vesting

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

type SmartContract struct {
	kalpsdk.Contract
}

func (s *SmartContract) Initialize(ctx kalpsdk.TransactionContextInterface, startTimestamp uint64) error {
	if startTimestamp == 0 {
		return ErrCannotBeZero
	}

	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusBadRequest, "failed to get client id", err)
	}

	if signer != kalpFoundation {
		return NewCustomError(http.StatusBadRequest, "only kalp foundation can intialize the contract", err)
	}

	beneficiaryJSON, err := ctx.GetState(kalpFoundationBeneficiaryKey)
	if err != nil {
		return fmt.Errorf("failed to get Beneficiary struct for %s, %v", kalpFoundationBeneficiaryKey, err)
	}

	if beneficiaryJSON != nil {
		return fmt.Errorf("Contract is already initialised as %w: %s", ErrBeneficiaryAlreadyExists, kalpFoundationBeneficiaryKey)
	}

	// Initialize different vesting periods
	validateNSetVesting(ctx, Team.String(), 30*12*24*60*60, startTimestamp, 30*24*24*60*60, ConvertGiniToWei(300000000), 0)
	validateNSetVesting(ctx, Foundation.String(), 0, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(220000000), 0)
	validateNSetVesting(ctx, AngelRound.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(20000000), 0)
	validateNSetVesting(ctx, SeedRound.String(), 30*10*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(40000000), 0)
	validateNSetVesting(ctx, PrivateRound1.String(), 30*12*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(140000000), 0)
	validateNSetVesting(ctx, PrivateRound2.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(60000000), 0)
	validateNSetVesting(ctx, Advisors.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(30000000), 0)
	validateNSetVesting(ctx, KOLRound.String(), 30*3*24*60*60, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(30000000), 25)
	validateNSetVesting(ctx, Marketing.String(), 30*1*24*60*60, startTimestamp, 30*18*24*60*60, ConvertGiniToWei(80000000), 10)
	validateNSetVesting(ctx, StakingRewards.String(), 30*3*24*60*60, startTimestamp, 30*24*24*60*60, ConvertGiniToWei(180000000), 0)
	validateNSetVesting(ctx, EcosystemReserve.String(), 0, startTimestamp, 30*150*24*60*60, ConvertGiniToWei(560000000), 2)
	validateNSetVesting(ctx, Airdrop.String(), 30*6*24*60*60, startTimestamp, 30*9*24*60*60, ConvertGiniToWei(80000000), 10)
	validateNSetVesting(ctx, LiquidityPool.String(), 0, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(200000000), 25)
	validateNSetVesting(ctx, PublicAllocation.String(), 30*3*24*60*60, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(60000000), 25)

	err = SetBeneficiary(ctx, kalpFoundationBeneficiaryKey, &Beneficiary{
		TotalAllocations: kalpFoundationTotalAllocations,
		ClaimedAmount:    kalpFoundationClaimedAmount,
	})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set beneficiaries", err)
	}

	userVestingList := UserVestings{EcosystemReserve.String()}
	err = SetUserVesting(ctx, kalpFoundationUserVestingKey, userVestingList)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set user vestings", err)
	}

	return nil
}

func (s *SmartContract) AddBeneficiaries(ctx kalpsdk.TransactionContextInterface, vestingID string, beneficiaries []string, amounts []string) error {
	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusBadRequest, "failed to get client id", err)
	}

	if signer != kalpFoundation {
		return NewCustomError(http.StatusBadRequest, "only kalp foundation can intialize the contract", err)
	}

	vesting, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return fmt.Errorf("unable to get vesting: %v", err)
	}

	if len(beneficiaries) == 0 {
		return ErrNoBeneficiaries
	}
	if len(beneficiaries) != len(amounts) {
		return NewCustomError(http.StatusBadRequest, fmt.Sprintf("%w: %d != %d", ErrArraysLengthMismatch, len(beneficiaries), len(amounts)), nil)
	}

	// Total allocation calculation
	totalAllocations := big.NewInt(0)
	for i := 0; i < len(beneficiaries); i++ {
		amount, ok := new(big.Int).SetString(amounts[i], 10)
		if !ok {
			return InvalidAmountError("beneficiary", beneficiaries[i], amounts[i])
		}

		err := addBeneficiary(ctx, vestingID, beneficiaries[i], amounts[i])
		if err != nil {
			return err
		}

		totalAllocations.Add(totalAllocations, amount)
	}

	vestingTotalSupply, ok := new(big.Int).SetString(vesting.TotalSupply, 10)
	if !ok {
		return InvalidAmountError("vestingTotalSupply", vestingID, vesting.TotalSupply)
	}

	if vestingTotalSupply.Cmp(totalAllocations) < 0 {
		return NewCustomError(http.StatusBadRequest, fmt.Sprintf("%w: vesting type %d", ErrTotalSupplyReached, vestingID), nil)
	}

	vestingTotalSupply.Sub(vestingTotalSupply, totalAllocations)

	EmitBeneficiariesAdded(ctx, vestingID, totalAllocations.String())

	return nil
}

func (s *SmartContract) SetGiniToken(ctx kalpsdk.TransactionContextInterface, tokenAddress string) error {
	_, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusBadRequest, "failed to get client id", err)
	}

	if IsContractAddressValid(tokenAddress) {
		return ErrInvalidContractAddress
	}

	giniTokenAddress, err := ctx.GetState("giniToken")
	if err != nil {
		return fmt.Errorf("failed to get gini token state: %v", err)
	}
	if giniTokenAddress != nil && string(giniTokenAddress) != "" {
		return ErrContractAddressAlreadySet
	}

	err = ctx.PutStateWithoutKYC("giniToken", []byte(tokenAddress))
	if err != nil {
		return fmt.Errorf("failed to set gini token: %v", err)
	}

	event := map[string]interface{}{
		"eventType": giniTokenEvent,
		"token":     tokenAddress,
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %v", err)
	}

	err = ctx.SetEvent(giniTokenEvent, eventBytes)
	if err != nil {
		return fmt.Errorf("failed to emit SetGiniToken event: %v", err)
	}

	return nil
}

func (s *SmartContract) CalculateClaimAmount(ctx kalpsdk.TransactionContextInterface, beneficiaryAddress, vestingID string) (string, error) {
	// Retrieve beneficiary details
	beneficiaryKey := fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiaryAddress)
	beneficiaryJSON, err := ctx.GetState(beneficiaryKey)
	if err != nil {
		return "0", fmt.Errorf("failed to get beneficiary state for vesting %s and address %s: %v", vestingID, beneficiaryAddress, err)
	}
	if beneficiaryJSON == nil {
		return "0", fmt.Errorf("beneficiary not found for vesting %s and address %s", vestingID, beneficiaryAddress)
	}

	var beneficiary Beneficiary
	err = json.Unmarshal(beneficiaryJSON, &beneficiary)
	if err != nil {
		return "0", fmt.Errorf("failed to unmarshal beneficiary data: %v", err)
	}

	vestingJSON, err := ctx.GetState(vestingID)
	if err != nil {
		return "0", fmt.Errorf("failed to get vesting state for ID %s: %v", vestingID, err)
	}
	if vestingJSON == nil {
		return "0", fmt.Errorf("vesting period not found for ID %s", vestingID)
	}

	var vesting VestingPeriod
	err = json.Unmarshal(vestingJSON, &vesting)
	if err != nil {
		return "0", fmt.Errorf("failed to unmarshal vesting data: %v", err)
	}

	beneficiaryClaimedAmount, ok := new(big.Int).SetString(beneficiary.ClaimedAmount, 10)
	if !ok {
		return "0", fmt.Errorf("invalid amount format for vestingTotalSupply %s", vestingID)
	}

	beneficiaryTotalAllocations, ok := new(big.Int).SetString(beneficiary.TotalAllocations, 10)
	if !ok {
		return "0", fmt.Errorf("invalid amount format for vestingTotalSupply %s", vestingID)
	}

	if beneficiaryClaimedAmount == beneficiaryTotalAllocations {
		return "0", nil
	}

	// Calculate initial unlock
	currentTime := time.Now().Unix()
	if uint64(currentTime) <= vesting.CliffStartTimestamp {
		return "0", nil
	}

	initialUnlock := CalculateInitialUnlock(beneficiaryTotalAllocations, vesting.TGE)

	// Calculate claimable amount
	claimableAmount := CalculateClaimableAmount(
		uint64(currentTime),
		beneficiaryTotalAllocations,
		vesting.StartTimestamp,
		vesting.Duration,
		initialUnlock,
	)

	claimAmount := new(big.Int)
	claimAmount.Add(claimableAmount, initialUnlock)
	claimAmount.Sub(claimAmount, beneficiaryClaimedAmount)

	// Validate claim amount does not exceed total allocations
	if claimAmount.Add(claimableAmount, initialUnlock).Cmp(beneficiaryTotalAllocations) > 0 {
		return "0", fmt.Errorf("claim amount exceeds vesting amount for vesting ID %s and beneficiary %s: claimAmount=%d, totalAllocations=%d",
			vestingID, beneficiaryAddress, claimAmount, beneficiary.TotalAllocations)
	}

	return claimAmount.String(), nil
}

func CalculateInitialUnlock(totalAllocations *big.Int, initialUnlockPercentage uint64) *big.Int {
	if initialUnlockPercentage == 0 {
		return big.NewInt(0)
	}

	percentage := big.NewInt(int64(initialUnlockPercentage))

	result := new(big.Int).Mul(totalAllocations, percentage)
	return result.Div(result, big.NewInt(100))
}

func CalculateClaimableAmount(
	timestamp uint64,
	totalAllocations *big.Int,
	startTimestamp,
	duration uint64,
	initialUnlock *big.Int,
) *big.Int {
	if timestamp < startTimestamp {
		return big.NewInt(0)
	}

	elapsedIntervals := (timestamp - startTimestamp) / claimInterval
	if elapsedIntervals == 0 {
		return big.NewInt(0)
	}

	// If the timestamp is beyond the total duration, return the remaining amount
	endTimestamp := startTimestamp + duration
	if timestamp > endTimestamp {
		return new(big.Int).Sub(totalAllocations, initialUnlock)
	}

	// Calculate claimable amount
	allocationsAfterUnlock := new(big.Int).Sub(totalAllocations, initialUnlock)

	elapsed := big.NewInt(int64(elapsedIntervals))
	durationBig := big.NewInt(int64(duration))
	claimable := new(big.Int).Mul(allocationsAfterUnlock, elapsed)
	claimable.Div(claimable, durationBig)

	return claimable
}

// func (s *SmartContract) GetVestingData(ctx kalpsdk.TransactionContextInterface, vestingID VestingType) (*VestingPeriod, *big.Int, error) {
// 	vestingData, err := GetVestingPeriod(ctx, vestingID.String())
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("unable to get vesting: %v", err)
// 	}

// 	// Get claimed amount from state
// 	claimedAmountBytes, err := ctx.GetState(fmt.Sprintf("claimed_amount_%s", vestingID))
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to get claimed amount for vestingID %s: %v", vestingID, err)
// 	}
// 	claimedAmount := big.NewInt(0)
// 	if claimedAmountBytes != nil {
// 		claimedAmount.SetString(string(claimedAmountBytes), 10)
// 	}

// 	return vestingData, claimedAmount, nil
// }

// GetClaimsAmountForAllVestings returns total claim amount, vesting IDs, and claimable amounts for all user's vestings
func (s *SmartContract) GetClaimsAmountForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*big.Int, []string, []*big.Int, error) {
	totalAmount := big.NewInt(0)

	// Get all vestings for the beneficiary
	userVestingList, err := GetUserVesting(ctx, fmt.Sprintf("uservesting_%s", beneficiary))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	amounts := make([]*big.Int, len(userVestingList))

	for i, vestingID := range userVestingList {
		claimAmount, err := s.CalculateClaimAmount(ctx, beneficiary, vestingID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to calculate claim amount for vestingID %s: %v", vestingID, err)
		}

		amountInInt, ok := new(big.Int).SetString(claimAmount, 10)
		if !ok {
			return nil, nil, nil, InvalidAmountError("vestingID", vestingID, claimAmount)
		}

		totalAmount.Add(totalAmount, amountInInt)
		amounts[i] = amountInInt
	}

	return totalAmount, userVestingList, amounts, nil
}

// GetVestingsDuration returns the vesting durations for all user's vestings
func (s *SmartContract) GetVestingsDuration(ctx kalpsdk.TransactionContextInterface, beneficiary string) ([]string, []uint64, error) {
	userVestingList, err := GetUserVesting(ctx, fmt.Sprintf("uservesting_%s", beneficiary))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	vestingDurations := make([]uint64, len(userVestingList))

	for i, vestingID := range userVestingList {
		vestingData, err := GetVestingPeriod(ctx, vestingID)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get vesting: %v", err)
		}

		vestingDurations[i] = vestingData.Duration
	}

	return userVestingList, vestingDurations, nil
}

// GetAllocationsForAllVestings returns total allocations for each vesting of the beneficiary
func (s *SmartContract) GetAllocationsForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) ([]string, []*big.Int, error) {
	// Get all vestings for the beneficiary
	userVestingList, err := GetUserVesting(ctx, fmt.Sprintf("uservesting_%s", beneficiary))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	totalAllocations := make([]*big.Int, len(userVestingList))

	for i, vestingID := range userVestingList {
		beneficiaryData, err := GetBeneficiary(ctx, fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiary))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
		}

		totalAllocations[i] = big.NewInt(0)
		totalAllocations[i].SetString(beneficiaryData.TotalAllocations, 10)
	}

	return userVestingList, totalAllocations, nil
}

// GetTotalClaims returns the total claimed amount for each vesting of the beneficiary
func (s *SmartContract) GetTotalClaims(ctx kalpsdk.TransactionContextInterface, beneficiary string) ([]string, []*big.Int, error) {
	// Get all vestings for the beneficiary
	userVestingList, err := GetUserVesting(ctx, fmt.Sprintf("uservesting_%s", beneficiary))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	totalClaims := make([]*big.Int, len(userVestingList))

	for i, vestingID := range userVestingList {
		claimedAmountBytes, err := ctx.GetState(fmt.Sprintf("claimed_amount_%s", vestingID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get claimed amount for vestingID %s: %v", vestingID, err)
		}
		totalClaims[i] = big.NewInt(0)
		totalClaims[i].SetString(string(claimedAmountBytes), 10)
	}

	return userVestingList, totalClaims, nil
}