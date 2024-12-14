package vesting

import (
	"encoding/json"
	"fmt"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

type VestingPeriodEvent struct {
	VestingID           string
	TotalSupply         string
	CliffStartTimestamp uint64
	StartTimestamp      uint64
	EndTimestamp        uint64
	TGE                 uint64
}

type BeneficiariesAddedEvent struct {
	VestingID        string
	TotalAllocations string
}

type ClaimEvent struct {
	User      string
	VestingID string
	Amount    string
}

func EmitVestingInitialized(ctx kalpsdk.TransactionContextInterface, vestingID string,
	cliffDuration,
	startTimestamp,
	duration uint64,
	totalSupply string,
	tge uint64) error {
	vestingPeriod := VestingPeriodEvent{
		VestingID:           vestingID,
		TotalSupply:         totalSupply,
		CliffStartTimestamp: startTimestamp,
		StartTimestamp:      startTimestamp + cliffDuration,
		EndTimestamp:        startTimestamp + duration + cliffDuration,
		TGE:                 tge,
	}
	vestingPeriodJSON, err := json.Marshal(vestingPeriod)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}

	err = ctx.SetEvent(VestingInitializedKey, vestingPeriodJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func EmitBeneficiariesAdded(ctx kalpsdk.TransactionContextInterface, vestingID string, totalAllocations string) error {
	beneficiary := BeneficiariesAddedEvent{
		VestingID:        vestingID,
		TotalAllocations: totalAllocations,
	}

	beneficiaryJSON, err := json.Marshal(beneficiary)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}

	err = ctx.SetEvent(BeneficiariesAddedKey, beneficiaryJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func EmitSetGiniToken(ctx kalpsdk.TransactionContextInterface, tokenAddress string) error {
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

func EmitClaim(ctx kalpsdk.TransactionContextInterface, user, vestingID, amount string) error {
	claim := ClaimEvent{
		User:      user,
		VestingID: vestingID,
		Amount:    amount,
	}

	claimJSON, err := json.Marshal(claim)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}

	err = ctx.SetEvent(ClaimKey, claimJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}
