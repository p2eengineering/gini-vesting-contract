package vesting

import (
	"encoding/json"
	"fmt"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

func EmitVestingInitialized(sdk kalpsdk.TransactionContextInterface, vestingPeriod VestingPeriodEvent) error {
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
