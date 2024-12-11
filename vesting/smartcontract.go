package vesting

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

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
	// validateNSetVesting(ctx, Team.String(), 30*12*24*60*60, startTimestamp, 30*24*24*60*60, ConvertGiniToWei(300000000), 0)
	// validateNSetVesting(ctx, Foundation.String(), 0, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(220000000), 0)
	// validateNSetVesting(ctx, AngelRound.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(20000000), 0)
	// validateNSetVesting(ctx, SeedRound.String(), 30*10*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(40000000), 0)
	// validateNSetVesting(ctx, PrivateRound1.String(), 30*12*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(140000000), 0)
	// validateNSetVesting(ctx, PrivateRound2.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(60000000), 0)
	// validateNSetVesting(ctx, Advisors.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(30000000), 0)
	// validateNSetVesting(ctx, KOLRound.String(), 30*3*24*60*60, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(30000000), 25)
	// validateNSetVesting(ctx, Marketing.String(), 30*1*24*60*60, startTimestamp, 30*18*24*60*60, ConvertGiniToWei(80000000), 10)
	// validateNSetVesting(ctx, StakingRewards.String(), 30*3*24*60*60, startTimestamp, 30*24*24*60*60, ConvertGiniToWei(180000000), 0)
	// validateNSetVesting(ctx, EcosystemReserve.String(), 0, startTimestamp, 30*150*24*60*60, ConvertGiniToWei(560000000), 2)
	// validateNSetVesting(ctx, Airdrop.String(), 30*6*24*60*60, startTimestamp, 30*9*24*60*60, ConvertGiniToWei(80000000), 10)
	// validateNSetVesting(ctx, LiquidityPool.String(), 0, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(200000000), 25)
	// validateNSetVesting(ctx, PublicAllocation.String(), 30*3*24*60*60, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(60000000), 25)

	validateNSetVesting(ctx, Team.String(), 2*60, startTimestamp, 12*60, ConvertGiniToWei(300000000), 0)
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

	err = SetBeneficiary(ctx, EcosystemReserve.String(), kalpFoundationKey, &Beneficiary{
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

	if !IsContractAddressValid(tokenAddress) {
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
	fmt.Println("input arguments ----------->", vestingID, beneficiaryAddress)
	beneficiary, err := GetBeneficiary(ctx, vestingID, beneficiaryAddress)
	if err != nil {
		return "0", fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
	}

	fmt.Println("beneficiary --------> ", beneficiary)

	vesting, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return "0", fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
	}

	fmt.Println("vestingPeriod --------> ", vesting)

	beneficiaryClaimedAmount, ok := new(big.Int).SetString(beneficiary.ClaimedAmount, 10)
	if !ok {
		return "0", fmt.Errorf("invalid amount format for vestingTotalSupply %s", vestingID)
	}

	fmt.Println("beneficiaryClaimedAmount --------> ", beneficiaryClaimedAmount)

	beneficiaryTotalAllocations, ok := new(big.Int).SetString(beneficiary.TotalAllocations, 10)
	if !ok {
		return "0", fmt.Errorf("invalid amount format for vestingTotalSupply %s", vestingID)
	}

	fmt.Println("beneficiaryTotalAllocations --------> ", beneficiaryTotalAllocations)

	if beneficiaryClaimedAmount == beneficiaryTotalAllocations {
		return "0", nil
	}

	// Calculate initial unlock
	currentTime, _ := ctx.GetTxTimestamp()
	fmt.Println("uint64(currentTime) <= vesting.CliffStartTimestamp --------> ", uint64(currentTime.Seconds), vesting.CliffStartTimestamp, uint64(currentTime.Seconds) <= vesting.CliffStartTimestamp)

	if uint64(currentTime.Seconds) <= vesting.CliffStartTimestamp {
		return "0", nil
	}

	initialUnlock := CalculateInitialUnlock(beneficiaryTotalAllocations, vesting.TGE)

	fmt.Println("initialUnlock --------> ", initialUnlock)

	// Calculate claimable amount
	claimableAmount := CalculateClaimableAmount(
		uint64(currentTime.Seconds),
		beneficiaryTotalAllocations,
		vesting.StartTimestamp,
		vesting.Duration,
		initialUnlock,
	)

	fmt.Println("claimableAmount --------> ", claimableAmount, beneficiaryClaimedAmount)

	claimAmount := new(big.Int)
	claimAmount.Add(claimableAmount, initialUnlock)
	claimAmount.Sub(claimAmount, beneficiaryClaimedAmount)

	fmt.Println("claimAmount --------> ", claimAmount.String())

	// Validate claim amount does not exceed total allocations
	claimAmountExceeds := new(big.Int)
	claimAmountExceeds.Add(claimAmountExceeds, claimAmount)
	claimAmountExceeds.Add(claimAmountExceeds, beneficiaryClaimedAmount)
	if claimAmountExceeds.Cmp(beneficiaryTotalAllocations) > 0 {
		return "0", fmt.Errorf("claim amount exceeds vesting amount for vesting ID %s and beneficiary %s: claimAmount=%d, totalAllocations=%d",
			vestingID, beneficiaryAddress, claimAmount, beneficiary.TotalAllocations)
	}

	fmt.Println("claimAmount --------> ", claimAmount.String())

	return claimAmount.String(), nil
}

func CalculateInitialUnlock(totalAllocations *big.Int, initialUnlockPercentage uint64) *big.Int {
	fmt.Println("input arguments ----------->", initialUnlockPercentage, totalAllocations)

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
	fmt.Println("arguments --------> ", timestamp, totalAllocations, startTimestamp, duration, initialUnlock)

	if timestamp < startTimestamp {
		return big.NewInt(0)
	}

	elapsedIntervals := (timestamp - startTimestamp) / claimInterval
	fmt.Println("elapsedIntervals --------> ", elapsedIntervals)

	if elapsedIntervals == 0 {
		return big.NewInt(0)
	}

	// If the timestamp is beyond the total duration, return the remaining amount
	endTimestamp := startTimestamp + duration
	fmt.Println("endTimestamp --------> ", endTimestamp)

	if timestamp > endTimestamp {
		return new(big.Int).Sub(totalAllocations, initialUnlock)
	}

	// Calculate claimable amount
	allocationsAfterUnlock := new(big.Int).Sub(totalAllocations, initialUnlock)

	fmt.Println("allocationsAfterUnlock --------> ", allocationsAfterUnlock)

	elapsed := big.NewInt(int64(elapsedIntervals))
	fmt.Println("elapsed --------> ", elapsed)

	durationBig := big.NewInt(int64(duration))
	fmt.Println("durationBig --------> ", durationBig)

	durationBig.Div(durationBig, big.NewInt(claimInterval))

	claimable := new(big.Int).Mul(allocationsAfterUnlock, elapsed)
	fmt.Println("claimable, durationBig --------> ", claimable, durationBig)

	claimable.Div(claimable, durationBig)
	// fmt.Println("claimable --------> ", claimable)

	// claimIntervalInBigInt := big.NewInt(claimInterval)
	// claimable.Div(claimable, claimIntervalInBigInt)

	fmt.Println("claimable --------> ", claimable)

	return claimable
}

func (s *SmartContract) GetVestingData(ctx kalpsdk.TransactionContextInterface, vestingID string) (*VestingData, error) {
	vestingData, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return nil, fmt.Errorf("unable to get vesting: %v", err)
	}

	// Get claimed amount from state
	claimedAmountBytes, err := ctx.GetState(fmt.Sprintf("total_claims_%s", vestingID))
	if err != nil {
		return nil, fmt.Errorf("failed to get claimed amount for vestingID %s: %v", vestingID, err)
	}
	claimedAmount := big.NewInt(0)
	if claimedAmountBytes != nil {
		claimedAmount.SetString(string(claimedAmountBytes), 10)
	}

	return &VestingData{
		VestingPeriod: vestingData,
		ClaimedAmount: claimedAmount.String(),
	}, nil
}

func (s *SmartContract) ClaimAll(ctx kalpsdk.TransactionContextInterface, beneficiary string) error {
	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return fmt.Errorf("failed to get vesting list: %v", err)
	}

	fmt.Println("userVestingList --------> ", userVestingList)

	totalClaimAmount := big.NewInt(0)

	for _, vestingID := range userVestingList {
		beneficiaryData, err := GetBeneficiary(ctx, vestingID, beneficiary)
		if err != nil {
			return fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
		}

		fmt.Println("beneficiaryData --------> ", beneficiaryData)

		amountToClaim, err := s.CalculateClaimAmount(ctx, beneficiary, vestingID)
		if err != nil {
			return fmt.Errorf("failed to calculate claim amount: %v", err)
		}

		fmt.Println("amountToClaim --------> ", amountToClaim)

		amountToClaimInInt, ok := new(big.Int).SetString(amountToClaim, 10)
		if !ok {
			return InvalidAmountError("vestingID", vestingID, amountToClaim)
		}

		if amountToClaimInInt.Cmp(big.NewInt(0)) == 0 {
			return ErrNothingToClaim
		}

		claimedAmount := big.NewInt(0)
		claimedAmount.SetString(beneficiaryData.ClaimedAmount, 10)

		fmt.Println("claimedAmount --------> ", claimedAmount)

		claimedAmount.Add(claimedAmount, amountToClaimInInt)

		fmt.Println("claimedAmount --------> ", claimedAmount)

		beneficiaryData.ClaimedAmount = claimedAmount.String()

		fmt.Println("beneficiaryData --------> ", beneficiaryData)

		// Save updated beneficiary data
		if err = SetBeneficiary(ctx, vestingID, beneficiary, beneficiaryData); err != nil {
			return NewCustomError(http.StatusInternalServerError, "failed to set beneficiaries", err)
		}

		totalClaimsKey := fmt.Sprintf("total_claims_%s", vestingID)
		totalClaimsBytes, err := ctx.GetState(totalClaimsKey)
		if err != nil {
			return fmt.Errorf("failed to retrieve total claims data: %v", err)
		}

		totalClaims := big.NewInt(0)
		if totalClaimsBytes != nil {
			totalClaims.SetString(string(totalClaimsBytes), 10)
		}
		totalClaims.Add(totalClaims, amountToClaimInInt)

		fmt.Println("totalClaims --------> ", totalClaims)

		if err := ctx.PutStateWithoutKYC(totalClaimsKey, []byte(totalClaims.String())); err != nil {
			return fmt.Errorf("failed to update total claims for vesting ID %s: %v", vestingID, err)
		}

		totalClaimAmount.Add(totalClaimAmount, amountToClaimInInt)
	}

	if totalClaimAmount.Cmp(big.NewInt(0)) == 0 {
		return ErrNothingToClaim
	}

	fmt.Println("totalClaimAmount --------> ", totalClaimAmount)

	totalClaimsForAllBytes, err := ctx.GetState("total_claims_for_all")
	if err != nil {
		return fmt.Errorf("failed to retrieve total claims for all vestings: %v", err)
	}

	totalClaimsForAll := big.NewInt(0)
	if totalClaimsForAllBytes != nil {
		totalClaimsForAll.SetString(string(totalClaimsForAllBytes), 10)
	}
	totalClaimsForAll.Add(totalClaimsForAll, totalClaimAmount)

	fmt.Println("totalClaimsForAll --------> ", totalClaimsForAll)

	if err := ctx.PutStateWithoutKYC("total_claims_for_all", []byte(totalClaimsForAll.String())); err != nil {
		return fmt.Errorf("failed to update total claims for all vestings: %v", err)
	}

	// Emit Claim event (can be implemented as needed in your system)

	// // Simulate transfer of tokens (in a real system, you would interact with a token contract or handle appropriately)
	// if err := s.TransferTokens(ctx, signer, amountToClaimInInt); err != nil {
	// 	return fmt.Errorf("failed to transfer tokens: %v", err)
	// }

	return nil
}

// GetClaimsAmountForAllVestings returns total claim amount, vesting IDs, and claimable amounts for all user's vestings
func (s *SmartContract) GetClaimsAmountForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*ClaimsWithAllVestings, error) {
	totalAmount := big.NewInt(0)

	// Get all vestings for the beneficiary
	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	amounts := make([]string, len(userVestingList))

	for i, vestingID := range userVestingList {
		claimAmount, err := s.CalculateClaimAmount(ctx, beneficiary, vestingID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate claim amount for vestingID %s: %v", vestingID, err)
		}

		amountInInt, ok := new(big.Int).SetString(claimAmount, 10)
		if !ok {
			return nil, InvalidAmountError("vestingID", vestingID, claimAmount)
		}

		totalAmount.Add(totalAmount, amountInInt)
		amounts[i] = claimAmount
	}

	return &ClaimsWithAllVestings{
		TotalAmount:  totalAmount.String(),
		UserVestings: userVestingList,
		Amounts:      amounts,
	}, nil
}

// GetVestingsDuration returns the vesting durations for all user's vestings
func (s *SmartContract) GetVestingsDuration(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*VestingDurationsData, error) {
	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	vestingDurations := make([]uint64, len(userVestingList))

	for i, vestingID := range userVestingList {
		vestingData, err := GetVestingPeriod(ctx, vestingID)
		if err != nil {
			return nil, fmt.Errorf("unable to get vesting: %v", err)
		}

		vestingDurations[i] = vestingData.Duration
	}

	return &VestingDurationsData{
		UserVestings:     userVestingList,
		VestingDurations: vestingDurations,
	}, nil
}

// GetAllocationsForAllVestings returns total allocations for each vesting of the beneficiary
func (s *SmartContract) GetAllocationsForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*AllocationsWithAllVestings, error) {
	// Get all vestings for the beneficiary
	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	totalAllocations := make([]string, len(userVestingList))

	for i, vestingID := range userVestingList {
		beneficiaryData, err := GetBeneficiary(ctx, vestingID, beneficiary)
		if err != nil {
			return nil, fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
		}

		totalAllocations[i] = beneficiaryData.TotalAllocations
	}

	return &AllocationsWithAllVestings{
		UserVestings:     userVestingList,
		TotalAllocations: totalAllocations,
	}, nil
}

func (s *SmartContract) GetUserVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*UserVestingsData, error) {
	// Get all vestings for the beneficiary
	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	return &UserVestingsData{
		UserVestings: userVestingList,
	}, nil
}

// GetTotalClaims returns the total claimed amount for each vesting of the beneficiary
func (s *SmartContract) GetTotalClaims(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*TotalClaimsWithAllVestings, error) {
	// Get all vestings for the beneficiary
	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	totalClaims := make([]string, len(userVestingList))

	for i, vestingID := range userVestingList {
		beneficiaryData, err := GetBeneficiary(ctx, vestingID, beneficiary)
		if err != nil {
			return nil, fmt.Errorf("failed to get claimed amount for vestingID %s: %v", vestingID, err)
		}

		totalClaims[i] = beneficiaryData.ClaimedAmount
	}

	return &TotalClaimsWithAllVestings{
		UserVestings: userVestingList,
		TotalClaims:  totalClaims,
	}, nil
}

func (s *SmartContract) Claim(ctx kalpsdk.TransactionContextInterface, vestingID string) error {
	// Retrieve beneficiary data
	fmt.Println("arguments --------> ", vestingID)

	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusBadRequest, "failed to get client id", err)
	}

	beneficiary, err := GetBeneficiary(ctx, vestingID, signer)
	if err != nil {
		return fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
	}

	fmt.Println("beneficiary --------> ", beneficiary, signer)

	// Retrieve vesting period data
	vesting, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return fmt.Errorf("unable to get vesting: %v", err)
	}

	fmt.Println("vesting period --------> ", vesting)

	// Check if beneficiary has already claimed all allocations
	claimedAmount := big.NewInt(0)
	claimedAmount.SetString(beneficiary.ClaimedAmount, 10)

	totalAllocations := big.NewInt(0)
	totalAllocations.SetString(beneficiary.TotalAllocations, 10)

	fmt.Println("claimedAmount --------> ", claimedAmount, totalAllocations)

	if claimedAmount.Cmp(totalAllocations) == 0 {
		return ErrNothingToClaimAsAlreadyClaimed
	}

	// Calculate amount to claim
	amountToClaim, err := s.CalculateClaimAmount(ctx, signer, vestingID)
	if err != nil {
		return fmt.Errorf("failed to calculate claim amount: %v", err)
	}

	fmt.Println("amountToClaim --------> ", amountToClaim)

	amountToClaimInInt, ok := new(big.Int).SetString(amountToClaim, 10)
	if !ok {
		return InvalidAmountError("vestingID", vestingID, amountToClaim)
	}

	if amountToClaimInInt.Cmp(big.NewInt(0)) == 0 {
		timeStamp, _ := ctx.GetTxTimestamp()
		fmt.Println("timeStamp --------> ", timeStamp)

		if vesting.StartTimestamp > uint64(timeStamp.Seconds) {
			return fmt.Errorf("vesting has not started yet for vesting ID %s", vestingID)
		} else {
			return ErrNothingToClaim
		}
	}

	// Update claimed amount for beneficiary
	claimedAmount.Add(claimedAmount, amountToClaimInInt)
	beneficiary.ClaimedAmount = claimedAmount.String()

	fmt.Println("claimedAmount --------> ", claimedAmount, beneficiary)

	// Save updated beneficiary data
	if err = SetBeneficiary(ctx, vestingID, signer, beneficiary); err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set beneficiaries", err)
	}

	fmt.Println("claimedAmount --------> ", claimedAmount)

	// Update total claims for the vesting ID
	totalClaimsKey := fmt.Sprintf("total_claims_%s", vestingID)
	totalClaimsBytes, err := ctx.GetState(totalClaimsKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total claims data: %v", err)
	}

	totalClaims := big.NewInt(0)
	if totalClaimsBytes != nil {
		totalClaims.SetString(string(totalClaimsBytes), 10)
	}
	fmt.Println("totalClaims before ---------->", totalClaims)

	totalClaims.Add(totalClaims, amountToClaimInInt)

	fmt.Println("totalClaims after ---------->", totalClaims)

	if err := ctx.PutStateWithoutKYC(totalClaimsKey, []byte(totalClaims.String())); err != nil {
		return fmt.Errorf("failed to update total claims for vesting ID %s: %v", vestingID, err)
	}

	// Update total claims across all vestings
	totalClaimsForAllBytes, err := ctx.GetState("total_claims_for_all")
	if err != nil {
		return fmt.Errorf("failed to retrieve total claims for all vestings: %v", err)
	}

	totalClaimsForAll := big.NewInt(0)
	if totalClaimsForAllBytes != nil {
		totalClaimsForAll.SetString(string(totalClaimsForAllBytes), 10)
	}
	fmt.Println("totalClaimsForAll before ---------->", totalClaimsForAll)

	totalClaimsForAll.Add(totalClaimsForAll, amountToClaimInInt)

	fmt.Println("totalClaimsForAll after ---------->", totalClaimsForAll)

	if err := ctx.PutStateWithoutKYC("total_claims_for_all", []byte(totalClaimsForAll.String())); err != nil {
		return fmt.Errorf("failed to update total claims for all vestings: %v", err)
	}

	// Emit Claim event (can be implemented as needed in your system)

	// // Simulate transfer of tokens (in a real system, you would interact with a token contract or handle appropriately)
	// if err := s.TransferTokens(ctx, signer, amountToClaimInInt); err != nil {
	// 	return fmt.Errorf("failed to transfer tokens: %v", err)
	// }

	return nil
}

// // TODO: transfer
// func (s *SmartContract) TransferTokens(ctx kalpsdk.TransactionContextInterface, recipient string, amount string) error {
// 	fmt.Printf("Transferred %s tokens to %s\n", amount, recipient)
// 	return nil
// }
