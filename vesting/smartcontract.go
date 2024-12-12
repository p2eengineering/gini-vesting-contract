package vesting

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

type SmartContract struct {
	kalpsdk.Contract
}

func (s *SmartContract) Initialize(ctx kalpsdk.TransactionContextInterface, startTimestamp uint64) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("Initialize Invoked.... with arguments ", startTimestamp)

	if startTimestamp == 0 {
		return ErrCannotBeZero
	}

	if err := IsSignerKalpFoundation(ctx); err != nil {
		return err
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

	// validateNSetVesting(ctx, Team.String(), 2*60, startTimestamp, 12*60, ConvertGiniToWei(300000000), 0)
	// validateNSetVesting(ctx, Foundation.String(), 0, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(220000000), 0)
	// validateNSetVesting(ctx, AngelRound.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(20000000), 0)
	// validateNSetVesting(ctx, SeedRound.String(), 30*10*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(40000000), 0)
	// validateNSetVesting(ctx, PrivateRound1.String(), 30*12*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(140000000), 0)
	// validateNSetVesting(ctx, PrivateRound2.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(60000000), 0)
	// validateNSetVesting(ctx, Advisors.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(30000000), 0)
	// validateNSetVesting(ctx, KOLRound.String(), 30*3*24*60*60, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(30000000), 25)
	// validateNSetVesting(ctx, Marketing.String(), 2*60, startTimestamp, 12*60, ConvertGiniToWei(80000000), 10)
	// validateNSetVesting(ctx, StakingRewards.String(), 30*3*24*60*60, startTimestamp, 30*24*24*60*60, ConvertGiniToWei(180000000), 0)
	// validateNSetVesting(ctx, EcosystemReserve.String(), 0, startTimestamp, 12*60, ConvertGiniToWei(560000000), 2)
	// validateNSetVesting(ctx, Airdrop.String(), 30*6*24*60*60, startTimestamp, 30*9*24*60*60, ConvertGiniToWei(80000000), 10)
	// validateNSetVesting(ctx, LiquidityPool.String(), 0, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(200000000), 25)
	// validateNSetVesting(ctx, PublicAllocation.String(), 30*3*24*60*60, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(60000000), 25)

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
	logger := kalpsdk.NewLogger()
	logger.Infoln("AddBeneficiaries Invoked.... with arguments ", vestingID, beneficiaries, amounts)

	if err := IsSignerKalpFoundation(ctx); err != nil {
		return err
	}

	vestingPeriod, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return fmt.Errorf("unable to get vesting: %v", err)
	}

	if len(beneficiaries) == 0 {
		return ErrNoBeneficiaries
	}
	if len(beneficiaries) != len(amounts) {
		return ErrArraysLengthMismatch(len(beneficiaries), len(amounts))
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

	vestingTotalSupply, ok := new(big.Int).SetString(vestingPeriod.TotalSupply, 10)
	if !ok {
		return InvalidAmountError("vestingTotalSupply", vestingID, vestingPeriod.TotalSupply)
	}

	if vestingTotalSupply.Cmp(totalAllocations) < 0 {
		return NewCustomError(http.StatusBadRequest, fmt.Sprintf("%w: vesting type %d", ErrTotalSupplyReached, vestingID), nil)
	}

	vestingTotalSupply.Sub(vestingTotalSupply, totalAllocations)

	EmitBeneficiariesAdded(ctx, vestingID, totalAllocations.String())

	return nil
}

func (s *SmartContract) SetGiniToken(ctx kalpsdk.TransactionContextInterface, tokenAddress string) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("SetGiniToken Invoked.... with arguments ", tokenAddress)

	if err := IsSignerKalpFoundation(ctx); err != nil {
		return err
	}

	if !IsContractAddressValid(tokenAddress) {
		return ErrInvalidContractAddress(tokenAddress)
	}

	_, err := GetGiniTokenAddress(ctx)
	if err != nil {
		return err
	}

	err = SetGiniTokenAddress(ctx, tokenAddress)
	if err != nil {
		return fmt.Errorf("failed to set gini token: %v", err)
	}

	event := map[string]interface{}{
		"token": tokenAddress,
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
	logger := kalpsdk.NewLogger()
	logger.Infoln("CalculateClaimAmount Invoked.... with arguments ", beneficiaryAddress, vestingID)

	beneficiary, err := GetBeneficiary(ctx, vestingID, beneficiaryAddress)
	if err != nil {
		return "0", fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
	}

	vestingPeriod, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return "0", fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
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
	currentTime, _ := ctx.GetTxTimestamp()

	if uint64(currentTime.Seconds) <= vestingPeriod.CliffStartTimestamp {
		return "0", nil
	}

	initialUnlock := calcInitialUnlock(beneficiaryTotalAllocations, vestingPeriod.TGE)

	// Calculate claimable amount
	claimableAmount := calcClaimableAmount(
		uint64(currentTime.Seconds),
		beneficiaryTotalAllocations,
		vestingPeriod.StartTimestamp,
		vestingPeriod.Duration,
		initialUnlock,
	)

	claimAmount := new(big.Int)
	claimAmount.Add(claimableAmount, initialUnlock)
	claimAmount.Sub(claimAmount, beneficiaryClaimedAmount)

	// Validate claim amount does not exceed total allocations
	claimAmountExceeds := new(big.Int).Set(claimAmount)
	claimAmountExceeds.Add(claimAmountExceeds, beneficiaryClaimedAmount)
	if claimAmountExceeds.Cmp(beneficiaryTotalAllocations) > 0 {
		return "0", fmt.Errorf("claim amount exceeds vesting amount for vesting ID %s and beneficiary %s: claimAmount=%d, totalAllocations=%d",
			vestingID, beneficiaryAddress, claimAmount, beneficiary.TotalAllocations)
	}

	logger.Infoln("CalculateClaimAmount Invoked complete.... with output ", claimAmount.String())

	return claimAmount.String(), nil
}

func (s *SmartContract) GetVestingData(ctx kalpsdk.TransactionContextInterface, vestingID string) (*VestingData, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetVestingData Invoked.... with arguments ", vestingID)

	vestingPeriod, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return nil, fmt.Errorf("unable to get vesting: %v", err)
	}

	claimedAmount, err := GetClaimedAmount(ctx, vestingID)
	if err != nil {
		return nil, err
	}

	logger.Infoln("GetVestingData Invoked complete.... with output ", vestingPeriod, claimedAmount.String())

	return &VestingData{
		VestingPeriod: vestingPeriod,
		ClaimedAmount: claimedAmount.String(),
	}, nil
}

func (s *SmartContract) ClaimAll(ctx kalpsdk.TransactionContextInterface, beneficiary string) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetVestingData Invoked.... with arguments ", beneficiary)
	
	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to get client id", err)
	}

	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return fmt.Errorf("failed to get vesting list: %v", err)
	}

	totalClaimAmount := big.NewInt(0)

	for _, vestingID := range userVestingList {
		beneficiaryData, err := GetBeneficiary(ctx, vestingID, beneficiary)
		if err != nil {
			return fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
		}

		amountToClaim, err := s.CalculateClaimAmount(ctx, beneficiary, vestingID)
		if err != nil {
			return fmt.Errorf("failed to calculate claim amount: %v", err)
		}

		amountToClaimInInt, ok := new(big.Int).SetString(amountToClaim, 10)
		if !ok {
			return InvalidAmountError("vestingID", vestingID, amountToClaim)
		}

		if amountToClaimInInt.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		claimedAmount := big.NewInt(0)
		claimedAmount.SetString(beneficiaryData.ClaimedAmount, 10)
		claimedAmount.Add(claimedAmount, amountToClaimInInt)

		beneficiaryData.ClaimedAmount = claimedAmount.String()

		// Save updated beneficiary data
		if err = SetBeneficiary(ctx, vestingID, beneficiary, beneficiaryData); err != nil {
			return NewCustomError(http.StatusInternalServerError, "failed to set beneficiaries", err)
		}

		totalClaims, err := GetTotalClaims(ctx, vestingID)
		if err != nil {
			return fmt.Errorf("failed to get total claims data: %v", err)
		}

		totalClaims.Add(totalClaims, amountToClaimInInt)

		if err := SetTotalClaims(ctx, vestingID, totalClaims); err != nil {
			return fmt.Errorf("failed to update total claims for vesting ID %s: %v", vestingID, err)
		}

		totalClaimAmount.Add(totalClaimAmount, amountToClaimInInt)
		EmitClaim(ctx, signer, vestingID, amountToClaim)
	}

	if totalClaimAmount.Cmp(big.NewInt(0)) == 0 {
		return ErrNothingToClaim
	}

	totalClaimsForAll, err := GetTotalClaimsForAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total claims for all vestings: %v", err)
	}

	totalClaimsForAll.Add(totalClaimsForAll, totalClaimAmount)

	if err := SetTotalClaimsForAll(ctx, totalClaimsForAll); err != nil {
		return fmt.Errorf("failed to update total claims for all vestings: %v", err)
	}

	// Simulate transfer of tokens (in a real system, you would interact with a token contract or handle appropriately)
	output := ctx.InvokeChaincode(crossCC_GINI_Contract, [][]byte{[]byte(crossCC_GINI_Transfer), []byte(signer), []byte(totalClaimAmount.String())}, Channel)

	b, _ := strconv.ParseBool(string(output.Payload))
	if !b {
		return NewCustomError(http.StatusInternalServerError, "unable to transfer token", err)
	}

	return nil
}

// GetClaimsAmountForAllVestings returns total claim amount, vesting IDs, and claimable amounts for all user's vestings
func (s *SmartContract) GetClaimsAmountForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*ClaimsWithAllVestings, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetClaimsAmountForAllVestings Invoked.... with arguments ", beneficiary)

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

	logger.Infoln("GetClaimsAmountForAllVestings Invoked complete.... with output ", totalAmount.String(), userVestingList, amounts)

	return &ClaimsWithAllVestings{
		TotalAmount:  totalAmount.String(),
		UserVestings: userVestingList,
		Amounts:      amounts,
	}, nil
}

// GetVestingsDuration returns the vesting durations for all user's vestings
func (s *SmartContract) GetVestingsDuration(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*VestingDurationsData, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetVestingsDuration Invoked.... with input arguments ", beneficiary)

	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	vestingDurations := make([]uint64, len(userVestingList))

	for i, vestingID := range userVestingList {
		vestingPeriod, err := GetVestingPeriod(ctx, vestingID)
		if err != nil {
			return nil, fmt.Errorf("unable to get vesting: %v", err)
		}

		vestingDurations[i] = vestingPeriod.Duration
	}

	logger.Infoln("GetVestingsDuration Invoked complete.... with output ", userVestingList, vestingDurations)

	return &VestingDurationsData{
		UserVestings:     userVestingList,
		VestingDurations: vestingDurations,
	}, nil
}

// GetAllocationsForAllVestings returns total allocations for each vesting of the beneficiary
func (s *SmartContract) GetAllocationsForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*AllocationsWithAllVestings, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetAllocationsForAllVestings Invoked.... with input arguments ", beneficiary)

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

	logger.Infoln("GetAllocationsForAllVestings Invoked complete.... with output ", userVestingList, totalAllocations)

	return &AllocationsWithAllVestings{
		UserVestings:     userVestingList,
		TotalAllocations: totalAllocations,
	}, nil
}

func (s *SmartContract) GetUserVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*UserVestingsData, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetUserVestings Invoked.... with arguments ", beneficiary)

	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	logger.Infoln("GetUserVestings Invoked complete.... with output ", userVestingList)

	return &UserVestingsData{
		UserVestings: userVestingList,
	}, nil
}

// GetTotalClaims returns the total claimed amount for each vesting of the beneficiary
func (s *SmartContract) GetTotalClaims(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*TotalClaimsWithAllVestings, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetTotalClaims Invoked.... with arguments ", beneficiary)

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

	logger.Infoln("GetTotalClaims Invoked complete.... with output ", userVestingList, totalClaims)

	return &TotalClaimsWithAllVestings{
		UserVestings: userVestingList,
		TotalClaims:  totalClaims,
	}, nil
}

func (s *SmartContract) Claim(ctx kalpsdk.TransactionContextInterface, vestingID string) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("Claim Invoked.... with arguments ", vestingID)

	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to get client id", err)
	}

	beneficiary, err := GetBeneficiary(ctx, vestingID, signer)
	if err != nil {
		return fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
	}

	// Retrieve vesting period data
	vestingPeriod, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return fmt.Errorf("unable to get vesting: %v", err)
	}

	// Check if beneficiary has already claimed all allocations
	claimedAmount := big.NewInt(0)
	claimedAmount.SetString(beneficiary.ClaimedAmount, 10)

	totalAllocations := big.NewInt(0)
	totalAllocations.SetString(beneficiary.TotalAllocations, 10)

	if claimedAmount.Cmp(totalAllocations) == 0 {
		return ErrNothingToClaim
	}

	// Calculate amount to claim
	amountToClaim, err := s.CalculateClaimAmount(ctx, signer, vestingID)
	if err != nil {
		return fmt.Errorf("failed to calculate claim amount: %v", err)
	}

	amountToClaimInInt, ok := new(big.Int).SetString(amountToClaim, 10)
	if !ok {
		return InvalidAmountError("vestingID", vestingID, amountToClaim)
	}

	if amountToClaimInInt.Cmp(big.NewInt(0)) == 0 {
		timeStamp, _ := ctx.GetTxTimestamp()

		if vestingPeriod.StartTimestamp > uint64(timeStamp.Seconds) {
			return OnlyAfterVestingStart(vestingID)
		} else {
			return ErrNothingToClaim
		}
	}

	// Update claimed amount for beneficiary
	claimedAmount.Add(claimedAmount, amountToClaimInInt)
	beneficiary.ClaimedAmount = claimedAmount.String()

	// Save updated beneficiary data
	if err = SetBeneficiary(ctx, vestingID, signer, beneficiary); err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set beneficiaries", err)
	}

	totalClaims, err := GetTotalClaims(ctx, vestingID)
	if err != nil {
		return fmt.Errorf("failed to get total claims data: %v", err)
	}

	totalClaims.Add(totalClaims, amountToClaimInInt)

	if err := SetTotalClaims(ctx, vestingID, totalClaims); err != nil {
		return fmt.Errorf("failed to update total claims for vesting ID %s: %v", vestingID, err)
	}

	totalClaimsForAll, err := GetTotalClaimsForAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get total claims for all vestings: %v", err)
	}

	totalClaimsForAll.Add(totalClaimsForAll, amountToClaimInInt)

	if err := SetTotalClaimsForAll(ctx, totalClaimsForAll); err != nil {
		return fmt.Errorf("failed to update total claims for all vestings: %v", err)
	}

	// Emit Claim event (can be implemented as needed in your system)
	EmitClaim(ctx, signer, vestingID, amountToClaim)

	// Simulate transfer of tokens (in a real system, you would interact with a token contract or handle appropriately)
	output := ctx.InvokeChaincode(crossCC_GINI_Contract, [][]byte{[]byte(crossCC_GINI_Transfer), []byte(signer), []byte(amountToClaim)}, Channel)

	b, _ := strconv.ParseBool(string(output.Payload))
	if !b {
		return NewCustomError(http.StatusInternalServerError, "unable to transfer token", err)
	}

	return nil
}
