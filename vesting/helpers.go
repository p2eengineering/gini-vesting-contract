package vesting

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

var (
	beneficiaries = make(map[string]map[string]*Beneficiary)
	userVestings  = make(map[string][]string)
)

func addBeneficiary(ctx kalpsdk.TransactionContextInterface, vestingID, beneficiary, amount string) error {
	// Ensure beneficiary is not zero address
	if beneficiary == "" {
		return errors.New("beneficiary address cannot be zero")
	}

	amountInInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return fmt.Errorf("invalid amount format for beneficiary %s", beneficiary)
	}

	// Ensure amount is not zero
	if amountInInt.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("%w: %s", ErrZeroVestingAmount, beneficiary)
	}

	beneficiaryJSON, err := ctx.GetState(fmt.Sprintf("beneficiary_%s_%s", vestingID, beneficiary))
	if err != nil {
		return fmt.Errorf("failed to get user vestings for %s: %v", beneficiary, err)
	}

	var beneficiaryStruct *Beneficiary
	// Debug existing state
	fmt.Printf("Existing state: %s\n", string(beneficiaryJSON))

	// Unmarshal existing data
	if beneficiaryJSON == nil {
		beneficiaryJSON, err = json.Marshal(&Beneficiary{
			TotalAllocations: amount,
			ClaimedAmount:    "0", // Initialize with zero
		})
		if err != nil {
			return fmt.Errorf("failed to marshal beneficiaries: %s", err.Error())
		}
	} else {
		err = json.Unmarshal(beneficiaryJSON, &beneficiaryStruct)
		if err != nil {
			return fmt.Errorf("failed to unmarshal user vesting list for %s: %v", beneficiary, err)
		}

		fmt.Printf("Existing beneficiaryStruct state: %s\n", beneficiaryStruct, *beneficiaryStruct)

		if beneficiaryStruct != nil {
			totalAllocationsInInt, ok := new(big.Int).SetString(beneficiaryStruct.TotalAllocations, 10)
			if !ok {
				return fmt.Errorf("invalid amount format for beneficiary %s", beneficiary)
			}

			if totalAllocationsInInt.Cmp(big.NewInt(0)) != 0 {
				return fmt.Errorf("%w: %s", ErrBeneficiaryAlreadyExists, beneficiary)
			}
		}
	}

	err = ctx.PutStateWithoutKYC(fmt.Sprintf("beneficiary_%s_%s", vestingID, beneficiary), beneficiaryJSON)
	if err != nil {
		return fmt.Errorf("failed to set vestingPeriod: %v", err)
	}

	fmt.Println("hello userveting----------->", fmt.Sprintf("uservesting_%s", beneficiary))

	userVestingJSON, err := ctx.GetState(fmt.Sprintf("uservesting_%s", beneficiary))
	if err != nil {
		return fmt.Errorf("failed to get user vestings for %s: %v", beneficiary, err)
	}

	fmt.Println("hello userveting json----------->", string(userVestingJSON))

	var userVestingList []string
	if userVestingJSON == nil {
		// Initialize if no data exists
		userVestingList = []string{}
	} else {
		// Debug existing state
		fmt.Printf("Existing state: %s\n", string(userVestingJSON))

		// Unmarshal existing data
		err = json.Unmarshal(userVestingJSON, &userVestingList)
		if err != nil {
			return fmt.Errorf("failed to unmarshal user vesting list for %s: %v", beneficiary, err)
		}
	}

	fmt.Printf("Existing array state: %s\n", userVestingList)

	// Append the new vestingID
	userVestingList = append(userVestingList, vestingID)

	// Marshal the updated list
	updatedUserVestingJSON, err := json.Marshal(userVestingList)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user vesting list for %s: %v", beneficiary, err)
	}

	// Debug the JSON being saved
	fmt.Printf("Saving data: Key=%s, Value=%s\n", fmt.Sprintf("uservesting_%s", beneficiary), string(updatedUserVestingJSON))

	// Save the updated state
	err = ctx.PutStateWithoutKYC(fmt.Sprintf("uservesting_%s", beneficiary), updatedUserVestingJSON)
	if err != nil {
		return fmt.Errorf("failed to set updated user vesting list for %s: %v", beneficiary, err)
	}

	fmt.Printf("Beneficiary %s added to vesting %s with allocation %s\n", beneficiary, vestingID, amount)
	return nil
}

// Function to get extract the userId from ca identity.  It is required to for checking the minter
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
