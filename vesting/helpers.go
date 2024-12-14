package vesting

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strings"

	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

func GetUserId(ctx kalpsdk.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
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
	isValid, _ := regexp.MatchString(contractAddressRegex, address)
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

func IsSignerKalpFoundation(ctx kalpsdk.TransactionContextInterface) error {
	signer, err := GetUserId(ctx)
	if err != nil {
		return NewCustomError(http.StatusInternalServerError, "failed to get client id", err)
	}

	if signer != kalpFoundation {
		return NewCustomError(http.StatusBadRequest, "signer is not kalp foundation", err)
	}

	return nil
}
