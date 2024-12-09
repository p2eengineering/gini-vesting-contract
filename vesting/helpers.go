package vesting

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

func addBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiary, amount string) error {
	// Ensure beneficiary is not zero address
	if IsUserAddressValid(beneficiary) {
		return errors.New("beneficiary address cannot be zero")
	}

	amountInInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return InvalidAmountError("beneficiary", beneficiary)
	}

	// Ensure amount is not zero
	if amountInInt.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("%w: %s", ErrZeroVestingAmount, beneficiary)
	}

	beneficiaryJSON, err := ctx.GetState(fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiary))
	if err != nil {
		return fmt.Errorf("failed to get Beneficiary struct for vestingId : %s and beneficiary: %s, %v", vestingID, beneficiary, err)
	}

	var beneficiaryStruct *Beneficiary

	if beneficiaryJSON == nil {
		beneficiaryJSON, err = json.Marshal(&Beneficiary{
			TotalAllocations: amount,
			ClaimedAmount:    "0",
		})
		if err != nil {
			return fmt.Errorf("failed to marshal beneficiaries: %s", err.Error())
		}
	} else {
		err = json.Unmarshal(beneficiaryJSON, &beneficiaryStruct)
		if err != nil {
			return fmt.Errorf("failed to unmarshal beneficiary for %s: %v", beneficiary, err)
		}

		if beneficiaryStruct != nil {
			totalAllocationsInInt, ok := new(big.Int).SetString(beneficiaryStruct.TotalAllocations, 10)
			if !ok {
				return InvalidAmountError("beneficiary", beneficiary)
			}

			if totalAllocationsInInt.Cmp(big.NewInt(0)) != 0 {
				return fmt.Errorf("%w: %s", ErrBeneficiaryAlreadyExists, beneficiary)
			}
		}
	}

	err = ctx.PutStateWithoutKYC(fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiary), beneficiaryJSON)
	if err != nil {
		return fmt.Errorf("failed to set vestingPeriod: %v", err)
	}

	userVestingList, err := GetUserVesting(ctx, fmt.Sprintf("uservesting_%s", beneficiary))
	if err != nil {
		return fmt.Errorf("failed to get vesting list: %v", err)
	}

	userVestingList = append(userVestingList, vestingID)

	err = SetUserVesting(ctx, fmt.Sprintf("uservesting_%s", beneficiary), userVestingList)
	if err != nil {
		return fmt.Errorf("failed to update vesting list: %v", err)
	}

	return nil
}

func GetUserId(sdk kalpsdk.TransactionContextInterface) (string, error) {
	b64ID, err := sdk.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to read clientID: %v", err)
	}

	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}

	completeId := string(decodeID)
	userId := completeId[(strings.Index(completeId, "x509::CN=") + 9):strings.Index(completeId, ",")]
	return userId, nil
}

func IsContractAddressValid(address string) bool {
	// Example validation logic (you can modify this to fit your use case)
	if address == "" {
		return false
	}
	// Assuming contract addresses should start with "0x" and have 42 characters
	isValid, _ := regexp.MatchString(hexAddressRegex, address)
	return isValid
}

// IsUserAddressValid validates if the user address is valid
func IsUserAddressValid(address string) bool {
	// Example validation logic (you can modify this to fit your use case)
	if address == "" {
		return false
	}
	// Assuming user addresses have the same structure as contract addresses
	isValid, _ := regexp.MatchString(hexAddressRegex, address)
	return isValid
}

func Decimals() uint64 {
	return 18 // You can modify this value if GINI uses a different decimal scheme
}

// ConvertGiniToWei converts a GINI token amount (uint64) to Wei (string)
func ConvertGiniToWei(giniAmount uint64) string {
	// Get the number of decimals for GINI (commonly 18)
	decimals := Decimals()

	// Convert giniAmount to a big.Int
	giniAmountBigInt := new(big.Int).SetUint64(giniAmount)

	// Calculate 10^decimals as the multiplier
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// Multiply GINI amount by 10^decimals to convert to Wei
	weiAmount := new(big.Int).Mul(giniAmountBigInt, multiplier)

	// Convert weiAmount to string and return
	return weiAmount.String()
}
