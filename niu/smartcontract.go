package vesting

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

// const kalpFoundation = "0b87970433b22494faff1cc7a819e71bddc7880c"
const kalpFoundation = "user1"
const kalpFoundationTotalAllocations = "560000000000000000000000000"
const kalpFoundationClaimedAmount = "11200000000000000000000000"
const kalpFoundationBeneficiary = "beneficiary_EcosystemReserve_kalp_foundation"
const kalpFoundationUserVesting = "uservesting_kalp_foundation"

const (
	Team             = "Team"
	Foundation       = "Foundation"
	AngelRound       = "AngelRound"
	SeedRound        = "SeedRound"
	PrivateRound1    = "PrivateRound1"
	PrivateRound2    = "PrivateRound2"
	Advisors         = "Advisors"
	KOLRound         = "KOLRound"
	Marketing        = "Marketing"
	StakingRewards   = "StakingRewards"
	EcosystemReserve = "EcosystemReserve"
	Airdrop          = "Airdrop"
	LiquidityPool    = "LiquidityPool"
	PublicAllocation = "PublicAllocation"
)

var vestingPeriods = make(map[string]*VestingPeriod)

type SmartContract struct {
	kalpsdk.Contract
}

func (s *SmartContract) Initialize(ctx kalpsdk.TransactionContextInterface, startTimestamp uint64) error {
	if startTimestamp == 0 {
		panic("CannotBeZero")
	}

	signer, err := GetUserId(ctx)
	if err != nil {
		return fmt.Errorf("error with status code %v, failed to get client id: %v", http.StatusBadRequest, err)
	}

	fmt.Println("signer ------------>", signer)
	if signer != kalpFoundation {
		return fmt.Errorf("error with status code %v, only kalp foundation can intialize the contract: %v", http.StatusBadRequest, err)
	}

	// Initialize different vesting periods
	ValidateNSetVesting(ctx, Team, 0, startTimestamp, 120, "300000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, Foundation, 0, startTimestamp, 30*12*24*60*60, "220000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, AngelRound, 30*6*24*60*60, startTimestamp, 30*12*24*60*60, "200000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, SeedRound, 30*10*24*60*60, startTimestamp, 30*12*24*60*60, "400000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, PrivateRound1, 30*12*24*60*60, startTimestamp, 30*12*24*60*60, "140000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, PrivateRound2, 30*6*24*60*60, startTimestamp, 30*12*24*60*60, "600000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, Advisors, 30*6*24*60*60, startTimestamp, 30*12*24*60*60, "300000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, KOLRound, 0, startTimestamp, 180, "300000000000000000000000000000000000", 25)
	ValidateNSetVesting(ctx, Marketing, 0, startTimestamp, 240, "800000000000000000000000000000000000", 10)
	ValidateNSetVesting(ctx, StakingRewards, 30*3*24*60*60, startTimestamp, 30*24*24*60*60, "180000000000000000000000000000000000", 0)
	ValidateNSetVesting(ctx, EcosystemReserve, 0, startTimestamp, 30*150*24*60*60, "560000000000000000000000000000000000", 2)
	ValidateNSetVesting(ctx, Airdrop, 60, startTimestamp, 1400, "800000000000000000000000000000000000", 10)
	ValidateNSetVesting(ctx, LiquidityPool, 0, startTimestamp, 30*6*24*60*60, "200000000000000000000000000000000000", 25)
	ValidateNSetVesting(ctx, PublicAllocation, 30*3*24*60*60, startTimestamp, 30*6*24*60*60, "600000000000000000000000000000000000", 25)

	fmt.Println("Initialize invoked done......", startTimestamp)
	for k, v := range vestingPeriods {
		fmt.Println("key, value", k, v)
	}

	beneficiaryJSON, err := json.Marshal(&Beneficiary{
		TotalAllocations: kalpFoundationTotalAllocations,
		ClaimedAmount:    kalpFoundationClaimedAmount, // Initialize with zero
	})
	if err != nil {
		return fmt.Errorf("failed to marshal beneficiaries: %s", err.Error())
	}

	err = ctx.PutStateWithoutKYC(kalpFoundationBeneficiary, beneficiaryJSON)
	if err != nil {
		return fmt.Errorf("failed to set vestingPeriod: %v", err)
	}

	userVestingList := []string{EcosystemReserve}

	// Marshal the updated list
	updatedUserVestingJSON, err := json.Marshal(userVestingList)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user vesting list for kalp foundation: %v", err)
	}

	err = ctx.PutStateWithoutKYC(kalpFoundationUserVesting, updatedUserVestingJSON)
	if err != nil {
		return fmt.Errorf("failed to set updated user vesting list for kalp foundation: %v", err)
	}

	return nil
}

func (s *SmartContract) AddBeneficiaries(ctx kalpsdk.TransactionContextInterface, vestingID string, beneficiaries []string, amounts []string) error {
	fmt.Println("AddBeneficiaries invoked......")

	vestingAsBytes, err := ctx.GetState(vestingID)
	if err != nil {
		return fmt.Errorf("vesting type %d does not exist", vestingID)
	}

	var vesting *VestingPeriod
	err = json.Unmarshal(vestingAsBytes, &vesting)
	if err != nil {
		return fmt.Errorf("failed to unmarshal vesting: %s", err.Error())
	}

	fmt.Println("vesting -------->", *vesting, vesting)

	// Check for empty arrays and length mismatch
	if len(beneficiaries) == 0 {
		return ErrNoBeneficiaries
	}
	if len(beneficiaries) != len(amounts) {
		return fmt.Errorf("%w: %d != %d", ErrArraysLengthMismatch, len(beneficiaries), len(amounts))
	}

	// Total allocation calculation
	totalAllocations := big.NewInt(0)
	for i := 0; i < len(beneficiaries); i++ {
		// Convert amount string to big.Int
		amount, ok := new(big.Int).SetString(amounts[i], 10)
		if !ok {
			return fmt.Errorf("invalid amount format for beneficiary %s", beneficiaries[i])
		}

		err := addBeneficiary(ctx, vestingID, beneficiaries[i], amounts[i])
		if err != nil {
			return err
		}

		fmt.Println("amounts -------->", totalAllocations, amount, *totalAllocations, *amount)

		// Add to totalAllocations
		totalAllocations.Add(totalAllocations, amount)
	}

	vestingTotalSupply, ok := new(big.Int).SetString(vesting.TotalSupply, 10)
	if !ok {
		return fmt.Errorf("invalid amount format for vestingTotalSupply %s", vestingID)
	}

	if vestingTotalSupply.Cmp(totalAllocations) < 0 {
		return fmt.Errorf("%w: vesting type %d", ErrTotalSupplyReached, vestingID)
	}

	// Subtract totalAllocations from vesting.TotalSupply
	vestingTotalSupply.Sub(vestingTotalSupply, totalAllocations)
	fmt.Printf("Beneficiaries added to vesting type %s. Total allocated: %s\n", vestingID, totalAllocations.String())

	return nil
}

func ValidateNSetVesting(
	ctx kalpsdk.TransactionContextInterface,
	vestingID string,
	cliffDuration uint64,
	startTimestamp uint64,
	duration uint64,
	totalSupply string,
	tge uint64,
) error {
	vestingPeriods[vestingID] = &VestingPeriod{
		TotalSupply:         totalSupply,
		CliffStartTimestamp: startTimestamp,
		StartTimestamp:      startTimestamp + cliffDuration,
		EndTimestamp:        startTimestamp + duration + cliffDuration,
		Duration:            duration,
		TGE:                 tge,
	}

	vestingPeriodsJSON, err := json.Marshal(vestingPeriods[vestingID])
	if err != nil {
		return fmt.Errorf("failed to marshal beneficiaries: %s", err.Error())
	}

	err = ctx.PutStateWithoutKYC(vestingID, vestingPeriodsJSON)
	if err != nil {
		return fmt.Errorf("failed to set vestingPeriod: %v", err)
	}

	// Emit Vesting Initialized event (simulate event using a print statement)
	EmitVestingInitialized(ctx, VestingPeriodEvent{
		VestingID:           vestingID,
		TotalSupply:         totalSupply,
		CliffStartTimestamp: startTimestamp,
		StartTimestamp:      startTimestamp + cliffDuration,
		EndTimestamp:        startTimestamp + duration + cliffDuration,
		TGE:                 tge,
	})

	return nil
}

// func (s *SmartContract) GetBeneficiaries(ctx kalpsdk.TransactionContextInterface, vestingID string) (map[string]*Beneficiary, error) {
// 	fmt.Println("GetBeneficiaries invoked...")

// 	// Fetch all beneficiaries for the given vestingID key from state
// 	beneficiariesAsBytes, err := ctx.GetState(vestingID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to read from world state: %s", err.Error())
// 	}

// 	if beneficiariesAsBytes == nil {
// 		return nil, fmt.Errorf("vesting ID %s does not exist", vestingID)
// 	}

// 	// Parse the beneficiaries map from JSON bytes
// 	beneficiaries := make(map[string]*Beneficiary)
// 	err = json.Unmarshal(beneficiariesAsBytes, &beneficiaries)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal beneficiaries: %s", err.Error())
// 	}

// 	return beneficiaries, nil
// }
