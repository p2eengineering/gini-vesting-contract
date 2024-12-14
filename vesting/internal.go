package vesting

import (
	"fmt"
	"math/big"

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

	err := SetVestingPeriod(ctx, vestingID, vestingPeriod)
	if err != nil {
		return fmt.Errorf("failed to set vestingPeriod: %v", err)
	}

	// Emit Vesting Initialized event (simulate event using a print statement)
	EmitVestingInitialized(ctx, vestingID, cliffDuration, startTimestamp, duration, totalSupply, tge)

	return nil
}

func addBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiary, amount string) error {
	if IsUserAddressValid(beneficiary) {
		return ErrInvalidUserAddress
	}

	amountInInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return InvalidAmountError("beneficiary", beneficiary, amount)
	}

	// Ensure amount is not zero
	if amountInInt.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("%w: %s", ErrNonPositiveVestingAmount, beneficiary)
	}

	beneficiaryJSON, err := ctx.GetState(fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiary))
	if err != nil {
		return fmt.Errorf("failed to get Beneficiary struct for vestingId : %s and beneficiary: %s, %v", vestingID, beneficiary, err)
	}

	if beneficiaryJSON != nil {
		return ErrBeneficiaryAlreadyExists(beneficiary)
	}

	err = SetBeneficiary(ctx, vestingID, beneficiary, &Beneficiary{
		TotalAllocations: amount,
		ClaimedAmount:    "0",
	})

	userVestingList, err := GetUserVesting(ctx, beneficiary)
	if err != nil {
		return fmt.Errorf("failed to get vesting list: %v", err)
	}

	userVestingList = append(userVestingList, vestingID)

	err = SetUserVesting(ctx, beneficiary, userVestingList)
	if err != nil {
		return fmt.Errorf("failed to update vesting list: %v", err)
	}

	return nil
}

func calcInitialUnlock(totalAllocations *big.Int, initialUnlockPercentage uint64) *big.Int {

	if initialUnlockPercentage == 0 {
		return big.NewInt(0)
	}

	percentage := big.NewInt(int64(initialUnlockPercentage))

	result := new(big.Int).Mul(totalAllocations, percentage)
	return result.Div(result, big.NewInt(100))
}

func calcClaimableAmount(
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
	durationBig.Div(durationBig, big.NewInt(claimInterval))
	claimable := new(big.Int).Mul(allocationsAfterUnlock, elapsed)
	claimable.Div(claimable, durationBig)

	return claimable
}
