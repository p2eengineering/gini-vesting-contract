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

	// Initialize different vesting periods
	validateNSetVesting(ctx, Team.String(), 0, startTimestamp, 120, ConvertGiniToWei(300000000000), 0)
	validateNSetVesting(ctx, Foundation.String(), 0, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(2200000000000), 0)
	validateNSetVesting(ctx, AngelRound.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(200000000000), 0)
	validateNSetVesting(ctx, SeedRound.String(), 30*10*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(400000000000), 0)
	validateNSetVesting(ctx, PrivateRound1.String(), 30*12*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(140000000000), 0)
	validateNSetVesting(ctx, PrivateRound2.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(600000000000), 0)
	validateNSetVesting(ctx, Advisors.String(), 30*6*24*60*60, startTimestamp, 30*12*24*60*60, ConvertGiniToWei(300000000000), 0)
	validateNSetVesting(ctx, KOLRound.String(), 0, startTimestamp, 180, ConvertGiniToWei(300000000000), 25)
	validateNSetVesting(ctx, Marketing.String(), 0, startTimestamp, 240, ConvertGiniToWei(800000000000), 10)
	validateNSetVesting(ctx, StakingRewards.String(), 30*3*24*60*60, startTimestamp, 30*24*24*60*60, ConvertGiniToWei(180000000000), 0)
	validateNSetVesting(ctx, EcosystemReserve.String(), 0, startTimestamp, 30*150*24*60*60, ConvertGiniToWei(560000000000), 2)
	validateNSetVesting(ctx, Airdrop.String(), 60, startTimestamp, 1400, ConvertGiniToWei(800000000000), 10)
	validateNSetVesting(ctx, LiquidityPool.String(), 0, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(200000000000), 25)
	validateNSetVesting(ctx, PublicAllocation.String(), 30*3*24*60*60, startTimestamp, 30*6*24*60*60, ConvertGiniToWei(600000000000), 25)

	err = SetBeneficiary(ctx, kalpFoundationBeneficiary, &Beneficiary{
		TotalAllocations: kalpFoundationTotalAllocations,
		ClaimedAmount:    kalpFoundationClaimedAmount,
	})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to set beneficiaries", err)
	}

	userVestingList := []string{EcosystemReserve.String()}
	err = SetUserVesting(ctx, kalpFoundationUserVesting, userVestingList)
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
			return InvalidAmountError("beneficiary", beneficiaries[i])
		}

		err := addBeneficiary(ctx, vestingID, beneficiaries[i], amounts[i])
		if err != nil {
			return err
		}

		totalAllocations.Add(totalAllocations, amount)
	}

	vestingTotalSupply, ok := new(big.Int).SetString(vesting.TotalSupply, 10)
	if !ok {
		return InvalidAmountError("vestingTotalSupply", vestingID)
	}

	if vestingTotalSupply.Cmp(totalAllocations) < 0 {
		return NewCustomError(http.StatusBadRequest, fmt.Sprintf("%w: vesting type %d", ErrTotalSupplyReached, vestingID), nil)
	}

	vestingTotalSupply.Sub(vestingTotalSupply, totalAllocations)

	EmitBeneficiariesAdded(ctx, vestingID, totalAllocations.String())
	
	return nil
}
