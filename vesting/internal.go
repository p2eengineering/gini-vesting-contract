package vesting

import (
	"encoding/json"
	"fmt"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

func validateNSetVesting(
	ctx kalpsdk.TransactionContextInterface,
	vestingID string,
	cliffDuration,
	startTimestamp,
	duration uint64,
	totalSupply string,
	tge uint64,
) error {
	vestingPeriod := &VestingPeriod{
		TotalSupply:         totalSupply,
		CliffStartTimestamp: startTimestamp,
		StartTimestamp:      startTimestamp + cliffDuration,
		EndTimestamp:        startTimestamp + duration + cliffDuration,
		Duration:            duration,
		TGE:                 tge,
	}

	vestingPeriodsJSON, err := json.Marshal(vestingPeriod)
	if err != nil {
		return fmt.Errorf("failed to marshal beneficiaries: %s", err.Error())
	}

	err = ctx.PutStateWithoutKYC(vestingID, vestingPeriodsJSON)
	if err != nil {
		return fmt.Errorf("failed to set vestingPeriod: %v", err)
	}

	// Emit Vesting Initialized event (simulate event using a print statement)
	EmitVestingInitialized(ctx, vestingID, cliffDuration, startTimestamp, duration, totalSupply, tge)

	return nil
}
