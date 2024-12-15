## While testing the contract whenever you are passing a bigInt pointer always make sure it is not modifying the value in that function

# TODO : we need to check if we add input validation for internal.go functions or not.
1. In validateNSetVesting lets add the validation for vestingID , startTimestamp > 0 , duration>0 , Add a TODO for totalSupply check
2. In calcInitialUnlock : should we add a check for totalAllocations
3. In calcClaimableAmount , Add validation timestamp > 0, TODO : Add a check for totalAllocations , startTimeStamp>0 , duration >0 ,initialUnlock>=0

# TODO : we might revert back to this later
1. ErrNoBeneficiaries = errors.New("no beneficiaries provided")
2. ErrCannotBeZero = errors.New("startTimestamp cannot be zero")
3. ErrInvalidUserAddress = errors.New("beneficiary address cannot be zero")

