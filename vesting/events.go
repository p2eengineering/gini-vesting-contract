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
	Signer        string
	VestingID     string
	AmountToClaim string
}

func EmitVestingInitialized(sdk kalpsdk.TransactionContextInterface, vestingID string,
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

	err = sdk.SetEvent("VestingInitialized", vestingPeriodJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func EmitBeneficiariesAdded(sdk kalpsdk.TransactionContextInterface, vestingID string, totalAllocations string) error {
	beneficiary := BeneficiariesAddedEvent{
		VestingID:        vestingID,
		TotalAllocations: totalAllocations,
	}

	beneficiaryJSON, err := json.Marshal(beneficiary)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}

	err = sdk.SetEvent("BeneficiariesAdded", beneficiaryJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func EmitClaim(sdk kalpsdk.TransactionContextInterface, signer, vestingID, amountToClaim string) error {
	claim := ClaimEvent{
		Signer:        signer,
		VestingID:     vestingID,
		AmountToClaim: amountToClaim,
	}

	claimJSON, err := json.Marshal(claim)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}

	err = sdk.SetEvent("Claim", claimJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}
