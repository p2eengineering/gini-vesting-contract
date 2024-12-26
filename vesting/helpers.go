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

	if !IsUserAddressValid(userId) {
		return "", ErrInvalidUserAddress(userId)
	}

	return userId, nil
}

func IsContractAddressValid(address string) bool {

	if address == "" {
		return false
	}

	isValid, _ := regexp.MatchString(contractAddressRegex, address)
	return isValid
}

func IsUserAddressValid(address string) bool {

	if address == "" {
		return false
	}

	isValid, _ := regexp.MatchString(hexAddressRegex, address)
	return isValid
}

func Decimals() uint64 {
	return 18
}

func ConvertGiniToWei(giniAmount uint64) string {
	decimals := Decimals()

	giniAmountBigInt := new(big.Int).SetUint64(giniAmount)

	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	weiAmount := new(big.Int).Mul(giniAmountBigInt, multiplier)

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

func isValidVestingID(vestingID string) bool {
	validVestingIDs := map[string]bool{
		"Team":             true,
		"Foundation":       true,
		"AngelRound":       true,
		"SeedRound":        true,
		"PrivateRound1":    true,
		"PrivateRound2":    true,
		"Advisors":         true,
		"KOLRound":         true,
		"Marketing":        true,
		"StakingRewards":   true,
		"EcosystemReserve": true,
		"Airdrop":          true,
		"LiquidityPool":    true,
		"PublicAllocation": true,
	}
	return validVestingIDs[vestingID]
}
