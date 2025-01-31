package vesting

import (
	"fmt"
	"math/big"
	"net/http"

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

	currentTime, err := ctx.GetTxTimestamp()
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "Failed to get transaction timestamp", nil)
	}

	if startTimestamp < uint64(currentTime.Seconds) {
		return ErrStartTimestampLessThanCurrentTimeStamp(startTimestamp, uint64(currentTime.Seconds))
	}

	if err := IsSignerKalpFoundation(ctx); err != nil {
		return err
	}

	kalpFoundationBeneficiaryKey, err := ctx.CreateCompositeKey(BeneficiariesPrefix, []string{EcosystemReserve.String(), kalpFoundation})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for kalpFoundationBeneficiary with vestingID %s and beneficiaryID with address %s", EcosystemReserve.String(), kalpFoundation), err)
	}

	beneficiaryJSON, err := ctx.GetState(kalpFoundationBeneficiaryKey)
	if err != nil {
		return fmt.Errorf("failed to get Beneficiary struct for %s, %v", kalpFoundation, err)
	}

	if beneficiaryJSON != nil {
		return fmt.Errorf("Contract is already initialised as %v", ErrBeneficiaryAlreadyExists(kalpFoundation))
	}

	kalpFoundationUserVestingKey, err := ctx.CreateCompositeKey(UserVestingsPrefix, []string{kalpFoundation})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for kalpFoundationUserVesting with beneficiaryID %s", kalpFoundation), err)
	}

	userVestingJSON, err := ctx.GetState(kalpFoundationUserVestingKey)
	if err != nil {
		return fmt.Errorf("failed to get User vesting struct for %s, %v", kalpFoundation, err)
	}

	if userVestingJSON != nil {
		return fmt.Errorf("Contract is already initialised as %v", ErrUserVestingsAlreadyExists(kalpFoundation))
	}

	validateNSetVesting(ctx, Team.String(), TeamCliffDuration, startTimestamp, TeamVestingDuration, ConvertGiniToWei(TeamTotalSupply), TeamTGE)
	validateNSetVesting(ctx, Foundation.String(), FoundationCliffDuration, startTimestamp, FoundationVestingDuration, ConvertGiniToWei(FoundationTotalSupply), FoundationTGE)
	validateNSetVesting(ctx, PrivateRound1.String(), PrivateRound1CliffDuration, startTimestamp, PrivateRound1VestingDuration, ConvertGiniToWei(PrivateRound1TotalSupply), PrivateRound1TGE)
	validateNSetVesting(ctx, PrivateRound2.String(), PrivateRound2CliffDuration, startTimestamp, PrivateRound2VestingDuration, ConvertGiniToWei(PrivateRound2TotalSupply), PrivateRound2TGE)
	validateNSetVesting(ctx, Advisors.String(), AdvisorsCliffDuration, startTimestamp, AdvisorsVestingDuration, ConvertGiniToWei(AdvisorsTotalSupply), AdvisorsTGE)
	validateNSetVesting(ctx, KOLRound.String(), KOLRoundCliffDuration, startTimestamp, KOLRoundVestingDuration, ConvertGiniToWei(KOLRoundTotalSupply), KOLRoundTGE)
	validateNSetVesting(ctx, Marketing.String(), MarketingCliffDuration, startTimestamp, MarketingVestingDuration, ConvertGiniToWei(MarketingTotalSupply), MarketingTGE)
	validateNSetVesting(ctx, StakingRewards.String(), StakingRewardsCliffDuration, startTimestamp, StakingRewardsVestingDuration, ConvertGiniToWei(StakingRewardsTotalSupply), StakingRewardsTGE)
	validateNSetVesting(ctx, EcosystemReserve.String(), EcosystemReserveCliffDuration, startTimestamp, EcosystemReserveVestingDuration, ConvertGiniToWei(EcosystemReserveTotalSupply), EcosystemReserveTGE)
	validateNSetVesting(ctx, Airdrop.String(), AirdropCliffDuration, startTimestamp, AirdropVestingDuration, ConvertGiniToWei(AirdropTotalSupply), AirdropTGE)
	validateNSetVesting(ctx, LiquidityPool.String(), LiquidityPoolCliffDuration, startTimestamp, LiquidityPoolVestingDuration, ConvertGiniToWei(LiquidityPoolTotalSupply), LiquidityPoolTGE)
	validateNSetVesting(ctx, PublicAllocation.String(), PublicAllocationCliffDuration, startTimestamp, PublicAllocationVestingDuration, ConvertGiniToWei(PublicAllocationTotalSupply), PublicAllocationTGE)

	err = SetBeneficiary(ctx, EcosystemReserve.String(), kalpFoundation, &Beneficiary{
		TotalAllocations: kalpFoundationTotalAllocations,
		ClaimedAmount:    kalpFoundationClaimedAmount,
	})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set beneficiaries", err)
	}

	EmitBeneficiariesAdded(ctx, EcosystemReserve.String(), kalpFoundationTotalAllocations)

	EmitClaim(ctx, kalpFoundation, EcosystemReserve.String(), kalpFoundationClaimedAmount)

	userVestingList := UserVestings{EcosystemReserve.String()}
	err = SetUserVesting(ctx, kalpFoundation, userVestingList)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set user vestings", err)
	}

	SetTotalSupplyForEcosystemReserve(ctx, EcosystemReserveCliffDuration, startTimestamp, EcosystemReserveVestingDuration, ConvertGiniToWei(EcosystemReserveTotalSupplyAfterInitialisation), EcosystemReserveTGE)

	return nil
}

func (s *SmartContract) AddBeneficiaries(ctx kalpsdk.TransactionContextInterface, vestingID string, beneficiaries []string, amounts []string) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("AddBeneficiaries Invoked.... with arguments ", vestingID, beneficiaries, amounts)

	if !isValidVestingID(vestingID) {
		return ErrInvalidVestingID(vestingID)
	}

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

	totalAllocations := big.NewInt(0)
	for i := 0; i < len(beneficiaries); i++ {
		amount, ok := new(big.Int).SetString(amounts[i], 10)
		if !ok {
			return ErrInvalidAmount("beneficiary", beneficiaries[i], amounts[i])
		}

		err := addBeneficiary(ctx, vestingID, beneficiaries[i], amounts[i])
		if err != nil {
			return err
		}

		totalAllocations.Add(totalAllocations, amount)
	}

	vestingTotalSupply, ok := new(big.Int).SetString(vestingPeriod.TotalSupply, 10)
	if !ok {
		return ErrInvalidAmount("vestingTotalSupply", vestingID, vestingPeriod.TotalSupply)
	}

	if vestingTotalSupply.Cmp(totalAllocations) < 0 {
		return ErrTotalSupplyReached(vestingID)
	}

	vestingTotalSupply.Sub(vestingTotalSupply, totalAllocations)

	vestingPeriod.TotalSupply = vestingTotalSupply.String()

	err = SetVestingPeriod(ctx, vestingID, vestingPeriod)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("unable to set vestingPeriod for vestingId : %s", vestingID), nil)
	}

	EmitBeneficiariesAdded(ctx, vestingID, totalAllocations.String())

	return nil
}

func (s *SmartContract) SetGiniToken(ctx kalpsdk.TransactionContextInterface, tokenAddress string) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("SetGiniToken Invoked.... with arguments ", tokenAddress)

	if err := IsSignerKalpFoundation(ctx); err != nil {
		return err
	}

	isContract, err := IsContractAddressValid(tokenAddress)
	if err != nil {
		return err
	}

	if !isContract {
		return ErrInvalidContractAddress(tokenAddress)
	}

	address, err := GetGiniTokenAddress(ctx)
	if err != nil {
		return err
	}

	if len(address) != 0 {
		return ErrTokenAlreadySet
	}

	err = SetGiniTokenAddress(ctx, tokenAddress)
	if err != nil {
		return fmt.Errorf("failed to set gini token: %v", err)
	}

	EmitSetGiniToken(ctx, tokenAddress)

	return nil
}

func (s *SmartContract) CalculateClaimAmount(ctx kalpsdk.TransactionContextInterface, beneficiaryAddress, vestingID string) (string, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("CalculateClaimAmount Invoked.... with arguments ", beneficiaryAddress, vestingID)

	if !isValidVestingID(vestingID) {
		return "0", ErrInvalidVestingID(vestingID)
	}

	isUser, err := IsUserAddressValid(beneficiaryAddress)
	if err != nil {
		return "0", err
	}

	if !isUser {
		return "0", ErrInvalidUserAddress(beneficiaryAddress)
	}

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

	currentTime, err := ctx.GetTxTimestamp()
	if err != nil {
		return "0", NewCustomError(http.StatusInternalServerError, "Failed to get transaction timestamp", nil)
	}

	if uint64(currentTime.Seconds) <= vestingPeriod.CliffStartTimestamp {
		return "0", nil
	}

	initialUnlock, err := calcInitialUnlock(beneficiaryTotalAllocations, vestingPeriod.TGE)
	if err != nil {
		return "0", err
	}

	claimableAmount, err := calcClaimableAmount(
		uint64(currentTime.Seconds),
		beneficiaryTotalAllocations,
		vestingPeriod.StartTimestamp,
		vestingPeriod.Duration,
		initialUnlock,
	)
	if err != nil {
		return "0", err
	}

	claimAmount := new(big.Int)
	claimAmount.Add(claimableAmount, initialUnlock)
	claimAmount.Sub(claimAmount, beneficiaryClaimedAmount)

	claimAmountExceeds := new(big.Int).Set(claimAmount)
	claimAmountExceeds.Add(claimAmountExceeds, beneficiaryClaimedAmount)
	if claimAmountExceeds.Cmp(beneficiaryTotalAllocations) > 0 {
		return "0", ErrClaimAmountExceedsVestingAmount(vestingID, beneficiaryAddress, claimAmount.String(), beneficiary.TotalAllocations)
	}

	logger.Infoln("CalculateClaimAmount Invoked complete.... with output ", claimAmount.String())

	return claimAmount.String(), nil
}

func (s *SmartContract) GetVestingData(ctx kalpsdk.TransactionContextInterface, vestingID string) (*VestingData, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetVestingData Invoked.... with arguments ", vestingID)

	if !isValidVestingID(vestingID) {
		return nil, ErrInvalidVestingID(vestingID)
	}

	vestingPeriod, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return nil, fmt.Errorf("unable to get vesting: %v", err)
	}

	claimedAmount, err := GetTotalClaims(ctx, vestingID)
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
	logger.Infoln("ClaimAll Invoked.... with arguments ", beneficiary)

	isUser, err := IsUserAddressValid(beneficiary)
	if err != nil {
		return err
	}

	if !isUser {
		return ErrInvalidUserAddress(beneficiary)
	}

	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to get client id", err)
	}

	if signer != beneficiary {
		return NewCustomError(http.StatusBadRequest, fmt.Sprintf("Signer '%s' does not match the beneficiary '%s'", signer, beneficiary), nil)
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
			return ErrInvalidAmount("vestingID", vestingID, amountToClaim)
		}

		if amountToClaimInInt.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		claimedAmount := big.NewInt(0)
		claimedAmount.SetString(beneficiaryData.ClaimedAmount, 10)
		claimedAmount.Add(claimedAmount, amountToClaimInInt)

		beneficiaryData.ClaimedAmount = claimedAmount.String()

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

	err = TransferGiniTokens(ctx, signer, totalClaimAmount.String())

	return err
}

func (s *SmartContract) GetClaimsAmountForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*ClaimsWithAllVestings, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetClaimsAmountForAllVestings Invoked.... with arguments ", beneficiary)

	isUser, err := IsUserAddressValid(beneficiary)
	if err != nil {
		return nil, err
	}

	if !isUser {
		return nil, ErrInvalidUserAddress(beneficiary)
	}

	totalAmount := big.NewInt(0)

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
			return nil, ErrInvalidAmount("vestingID", vestingID, claimAmount)
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

func (s *SmartContract) GetVestingsDuration(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*VestingDurationsData, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetVestingsDuration Invoked.... with input arguments ", beneficiary)

	isUser, err := IsUserAddressValid(beneficiary)
	if err != nil {
		return nil, err
	}

	if !isUser {
		return nil, ErrInvalidUserAddress(beneficiary)
	}

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

func (s *SmartContract) GetAllocationsForAllVestings(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*AllocationsWithAllVestings, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetAllocationsForAllVestings Invoked.... with input arguments ", beneficiary)

	isUser, err := IsUserAddressValid(beneficiary)
	if err != nil {
		return nil, err
	}

	if !isUser {
		return nil, ErrInvalidUserAddress(beneficiary)
	}

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

	isUser, err := IsUserAddressValid(beneficiary)
	if err != nil {
		return nil, err
	}

	if !isUser {
		return nil, ErrInvalidUserAddress(beneficiary)
	}

	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return nil, fmt.Errorf("failed to get vesting list: %v", err)
	}

	logger.Infoln("GetUserVestings Invoked complete.... with output ", userVestingList)

	return &UserVestingsData{
		UserVestings: userVestingList,
	}, nil
}

func (s *SmartContract) GetTotalClaims(ctx kalpsdk.TransactionContextInterface, beneficiary string) (*TotalClaimsWithAllVestings, error) {
	logger := kalpsdk.NewLogger()
	logger.Infoln("GetTotalClaims Invoked.... with arguments ", beneficiary)

	isUser, err := IsUserAddressValid(beneficiary)
	if err != nil {
		return nil, err
	}

	if !isUser {
		return nil, ErrInvalidUserAddress(beneficiary)
	}

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

	if !isValidVestingID(vestingID) {
		return ErrInvalidVestingID(vestingID)
	}

	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to get client id", err)
	}

	beneficiary, err := GetBeneficiary(ctx, vestingID, signer)
	if err != nil {
		return fmt.Errorf("failed to get beneficiary data for vestingID %s: %v", vestingID, err)
	}

	vestingPeriod, err := GetVestingPeriod(ctx, vestingID)
	if err != nil {
		return fmt.Errorf("unable to get vesting: %v", err)
	}

	claimedAmount := big.NewInt(0)
	claimedAmount.SetString(beneficiary.ClaimedAmount, 10)

	totalAllocations := big.NewInt(0)
	totalAllocations.SetString(beneficiary.TotalAllocations, 10)

	if claimedAmount.Cmp(totalAllocations) == 0 {
		return ErrNothingToClaim
	}

	amountToClaim, err := s.CalculateClaimAmount(ctx, signer, vestingID)
	if err != nil {
		return fmt.Errorf("failed to calculate claim amount: %v", err)
	}

	amountToClaimInInt, ok := new(big.Int).SetString(amountToClaim, 10)
	if !ok {
		return ErrInvalidAmount("vestingID", vestingID, amountToClaim)
	}

	if amountToClaimInInt.Cmp(big.NewInt(0)) == 0 {
		timeStamp, err := ctx.GetTxTimestamp()
		if err != nil {
			return NewCustomError(http.StatusInternalServerError, "Failed to get transaction timestamp", nil)
		}

		if vestingPeriod.StartTimestamp > uint64(timeStamp.Seconds) {
			return ErrOnlyAfterVestingStart(vestingID)
		} else {
			return ErrNothingToClaim
		}
	}

	claimedAmount.Add(claimedAmount, amountToClaimInInt)
	beneficiary.ClaimedAmount = claimedAmount.String()

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

	EmitClaim(ctx, signer, vestingID, amountToClaim)

	err = TransferGiniTokens(ctx, signer, amountToClaim)

	return err
}
