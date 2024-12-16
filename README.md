## While testing the contract whenever you are passing a bigInt pointer always make sure it is not modifying the value in that function

# TODO : we need to check if we add input validation for internal.go functions or not.
1. In validateNSetVesting lets add the validation for vestingID , startTimestamp > 0 , duration>0 , Add a TODO for totalSupply check
2. In calcInitialUnlock : should we add a check for totalAllocations
3. In calcClaimableAmount , Add validation timestamp > 0, TODO : Add a check for totalAllocations , startTimeStamp>0 , duration>0 ,initialUnlock>=0
4. Add the proper error for this scenario if the claim is called before TGE -> it should throw TGE not started
5. 

# TODO : we might revert back to this later
1. ErrNoBeneficiaries = errors.New("no beneficiaries provided")
2. ErrCannotBeZero = errors.New("startTimestamp cannot be zero")
3. ErrInvalidUserAddress = errors.New("beneficiary address cannot be zero")


Use this function for logging

func calcClaimableAmount(
	timestamp uint64,
	totalAllocations *big.Int,
	startTimestamp,
	duration uint64,
	initialUnlock *big.Int,
) *big.Int {

	fmt.Println("arguments in calcClaimableAmount", timestamp,
		totalAllocations,
		startTimestamp,
		duration,
		initialUnlock)

	if timestamp < startTimestamp {
		return big.NewInt(0)
	}

	elapsedIntervals := (timestamp - startTimestamp) / claimInterval

	fmt.Println("elapsed intervals", elapsedIntervals)
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
	claimable := new(big.Int).Div(allocationsAfterUnlock, durationBig)
	claimable.Mul(claimable, elapsed)

	return claimable
}

func TransferGiniTokens(ctx kalpsdk.TransactionContextInterface, signer, totalClaimAmount string) error {
	logger := kalpsdk.NewLogger()
	logger.Infoln("TransferGiniTokens called.... with arguments ", signer, totalClaimAmount)

	return nil

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

	// TODO: check this on stagenet also
	// Simulate transfer of tokens (in a real system, you would interact with a token contract or handle appropriately)
	output := ctx.InvokeChaincode(giniContract, [][]byte{[]byte(giniTransfer), []byte(signer), []byte(totalClaimAmount)}, channel)

	b, _ := strconv.ParseBool(string(output.Payload))

	if !b {
		return NewCustomError(int(output.Status), fmt.Sprintf("unable to transfer token: %s", output.Message), nil)
	}

	return nil
}
