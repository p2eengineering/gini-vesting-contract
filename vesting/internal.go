package vesting

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

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
	if !isValidVestingID(vestingID) {
		return ErrInvalidVestingID(vestingID)
	}

	if startTimestamp == 0 {
		return ErrCannotBeZero
	}

	if duration == 0 {
		return ErrDurationCannotBeZero(vestingID)
	}

	totalSupplyInInt, ok := new(big.Int).SetString(totalSupply, 10)
	if !ok {
		return ErrInvalidAmount("vestingID", vestingID, totalSupply)
	}

	if totalSupplyInInt.Cmp(big.NewInt(0)) <= 0 {
		return ErrTotalSupplyCannotBeNonPositive(vestingID)
	}

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

	EmitVestingInitialized(ctx, vestingID, cliffDuration, startTimestamp, duration, totalSupply, tge)

	return nil
}

func SetTotalSupplyForEcosystemReserve(
	ctx kalpsdk.TransactionContextInterface,
	cliffDuration,
	startTimestamp,
	duration uint64,
	totalSupply string,
	tge uint64,
) error {
	if startTimestamp == 0 {
		return ErrCannotBeZero
	}

	if duration == 0 {
		return ErrDurationCannotBeZero(EcosystemReserve.String())
	}

	totalSupplyInInt, ok := new(big.Int).SetString(totalSupply, 10)
	if !ok {
		return ErrInvalidAmount("vestingID", EcosystemReserve.String(), totalSupply)
	}

	if totalSupplyInInt.Cmp(big.NewInt(0)) < 0 {
		return ErrTotalSupplyCannotBeNegative(EcosystemReserve.String())
	}

	vestingPeriod := &VestingPeriod{
		TotalSupply:         totalSupply,
		CliffStartTimestamp: startTimestamp,
		StartTimestamp:      startTimestamp + cliffDuration,
		EndTimestamp:        startTimestamp + duration + cliffDuration,
		Duration:            duration,
		TGE:                 tge,
	}

	err := SetVestingPeriod(ctx, EcosystemReserve.String(), vestingPeriod)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to set vestingPeriod for VestingId: %s", EcosystemReserve.String()), nil)
	}

	EmitEventEcosystemReserveTotalSupplyChanged(ctx, kalpFoundationTotalAllocations)

	return nil
}

func addBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiary, amount string) error {
	isUser, err := IsUserAddressValid(beneficiary)
	if err != nil {
		return err
	}

	if !isUser {
		return ErrInvalidUserAddress(beneficiary)
	}

	amountInInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return ErrInvalidAmount("beneficiary", beneficiary, amount)
	}

	if amountInInt.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("%w: %s", ErrNonPositiveVestingAmount, beneficiary)
	}

	beneficiaryKey, err := ctx.CreateCompositeKey(BeneficiariesPrefix, []string{vestingID, beneficiary})
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to create the composite key for getting beneficiary with vestingID %s and beneficiaryID with address %s", vestingID, beneficiary), err)
	}

	beneficiaryJSON, err := ctx.GetState(beneficiaryKey)
	if err != nil {
		return fmt.Errorf("failed to get Beneficiary struct for vestingID : %s and beneficiary: %s, %v", vestingID, beneficiary, err)
	}

	if beneficiaryJSON != nil {
		return ErrBeneficiaryAlreadyExists(beneficiary)
	}

	err = SetBeneficiary(ctx, vestingID, beneficiary, &Beneficiary{
		TotalAllocations: amount,
		ClaimedAmount:    "0",
	})
	if err != nil {
		return err
	}

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

func calcInitialUnlock(totalAllocations *big.Int, initialUnlockPercentage uint64) (*big.Int, error) {

	if totalAllocations.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0), ErrTotalAllocationCannotBeNonPositive
	}

	if initialUnlockPercentage == 0 {
		return big.NewInt(0), nil
	}

	percentage := big.NewInt(int64(initialUnlockPercentage))

	result := new(big.Int).Mul(totalAllocations, percentage)
	return result.Div(result, big.NewInt(100)), nil
}

func calcClaimableAmount(
	timestamp uint64,
	totalAllocations *big.Int,
	startTimestamp,
	duration uint64,
	initialUnlock *big.Int,
) (*big.Int, error) {
	if timestamp == 0 {
		return big.NewInt(0), ErrCannotBeZero
	}

	if startTimestamp == 0 {
		return big.NewInt(0), ErrCannotBeZero
	}

	if duration == 0 {
		return big.NewInt(0), ErrDurationCannotBeZeroForClaimAmount
	}

	if totalAllocations.Cmp(big.NewInt(0)) <= 0 {
		return big.NewInt(0), ErrTotalAllocationCannotBeNonPositive
	}

	if initialUnlock.Cmp(big.NewInt(0)) < 0 {
		return big.NewInt(0), ErrInitialUnlockCannotBeNegative
	}

	if timestamp < startTimestamp {
		return big.NewInt(0), nil
	}

	elapsedIntervals := (timestamp - startTimestamp) / claimInterval

	if elapsedIntervals == 0 {
		return big.NewInt(0), nil
	}

	endTimestamp := startTimestamp + duration

	if timestamp > endTimestamp {
		return new(big.Int).Sub(totalAllocations, initialUnlock), nil
	}

	allocationsAfterUnlock := new(big.Int).Sub(totalAllocations, initialUnlock)

	elapsed := big.NewInt(int64(elapsedIntervals))
	durationBig := big.NewInt(int64(duration))
	durationBig.Div(durationBig, big.NewInt(claimInterval))
	claimable := new(big.Int).Div(allocationsAfterUnlock, durationBig)
	claimable.Mul(claimable, elapsed)

	return claimable, nil
}

func TransferGiniTokens(ctx kalpsdk.TransactionContextInterface, signer, totalClaimAmount string) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("TransferGiniTokens called.... with arguments ", signer, totalClaimAmount)

	giniContract, err := GetGiniTokenAddress(ctx)
	if err != nil {
		return err
	}

	if len(giniContract) == 0 {
		return NewCustomError(http.StatusNotFound, fmt.Sprintf("Gini token address with Key %s does not exist", giniTokenKey), nil)
	}

	channel := ctx.GetChannelID()
	if channel == "" {
		return NewCustomError(http.StatusInternalServerError, "unable to get the channel name", nil)
	}

	output := ctx.InvokeChaincode(giniContract, [][]byte{[]byte(giniTransfer), []byte(signer), []byte(totalClaimAmount)}, channel)

	b, err := strconv.ParseBool(string(output.Payload))
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, fmt.Sprintf("failed to parse output payload to boolean: %v", err), nil)
	}

	if !b {
		return NewCustomError(int(output.Status), fmt.Sprintf("unable to transfer token: %s", output.Message), nil)
	}

	return nil
}
