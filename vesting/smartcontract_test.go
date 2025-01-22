package vesting_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/p2eengineering/gini-vesting-contract/vesting"
	"github.com/p2eengineering/gini-vesting-contract/vesting/mocks"
	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
	"github.com/p2eengineering/kalp-sdk-public/response"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type transactionContext interface {
	kalpsdk.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	kalpsdk.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	kalpsdk.StateQueryIteratorInterface
}

//go:generate counterfeiter -o mocks/clientidentity.go -fake-name ClientIdentity . clientIdentity
type clientIdentity interface {
	cid.ClientIdentity
}

func SetUserID(transactionContext *mocks.TransactionContext, userID string) {
	completeId := fmt.Sprintf("x509::CN=%s,O=Organization,L=City,ST=State,C=Country", userID)

	// Base64 encode the complete ID
	b64ID := base64.StdEncoding.EncodeToString([]byte(completeId))

	clientIdentity := &mocks.ClientIdentity{}
	clientIdentity.GetIDReturns(b64ID, nil)
	transactionContext.GetClientIdentityReturns(clientIdentity)
}

func TestInitialize(t *testing.T) {
	t.Parallel()
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}
	// ****************START define helper functions*********************
	worldState := map[string][]byte{}
	transactionContext.CreateCompositeKeyStub = func(s1 string, s2 []string) (string, error) {
		key := "_" + s1 + "_"
		for _, s := range s2 {
			key += s + "_"
		}
		return key, nil
	}
	transactionContext.PutStateWithoutKYCStub = func(s string, b []byte) error {
		worldState[s] = b
		return nil
	}
	transactionContext.GetQueryResultStub = func(s string) (kalpsdk.StateQueryIteratorInterface, error) {
		var docType string
		var account string

		// finding doc type
		re := regexp.MustCompile(`"docType"\s*:\s*"([^"]+)"`)
		match := re.FindStringSubmatch(s)

		if len(match) > 1 {
			docType = match[1]
		}

		// finding account
		re = regexp.MustCompile(`"account"\s*:\s*"([^"]+)"`)
		match = re.FindStringSubmatch(s)

		if len(match) > 1 {
			account = match[1]
		}

		iteratorData := struct {
			index int
			data  []queryresult.KV
		}{}
		for key, val := range worldState {
			if strings.Contains(key, docType) && strings.Contains(key, account) {
				iteratorData.data = append(iteratorData.data, queryresult.KV{Key: key, Value: val})
			}
		}
		iterator := &mocks.StateQueryIterator{}
		iterator.HasNextStub = func() bool {
			return iteratorData.index < len(iteratorData.data)
		}
		iterator.NextStub = func() (*queryresult.KV, error) {
			if iteratorData.index < len(iteratorData.data) {
				iteratorData.index++
				return &iteratorData.data[iteratorData.index-1], nil
			}
			return nil, fmt.Errorf("iterator out of bounds")
		}
		return iterator, nil
	}
	transactionContext.GetStateStub = func(s string) ([]byte, error) {
		data, found := worldState[s]
		if found {
			return data, nil
		}
		return nil, nil
	}
	transactionContext.DelStateWithoutKYCStub = func(s string) error {
		delete(worldState, s)
		return nil
	}
	transactionContext.GetTxIDStub = func() string {
		const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
		length := 10
		rand.Seed(time.Now().UnixNano()) // Seed the random number generator
		result := make([]byte, length)
		for i := range result {
			result[i] = charset[rand.Intn(len(charset))]
		}
		return string(result)
	}
	// ****************END define helper functions*********************

	SetUserID(transactionContext, vesting.KalpFoundation)
	// transactionContext.GetKYCReturns(true, nil)

	err := vestingContract.Initialize(transactionContext, 199999999)
	require.NoError(t, err)

	// Test case for the invalid startTimestamp (0)
	err = vestingContract.Initialize(transactionContext, 0) // Passing 0 should trigger the error
	require.Error(t, err)
	require.Contains(t, err.Error(), "CannotBeZero")

	KalpFoundation := "0b87970433b22494faff1cc7a819e71bddc7880c"
	KalpFoundationBeneficiaryKeyPrefix := "beneficiaries_EcosystemReserve_"
	KalpFoundationUserVestingKeyPrefix := "uservestings_"
	kalpFoundationBeneficiaryKey := KalpFoundationBeneficiaryKeyPrefix + KalpFoundation
	kalpFoundationUserVestingKey := KalpFoundationUserVestingKeyPrefix + KalpFoundation

	beneficiaryJSON, err1 := transactionContext.GetStateStub(kalpFoundationBeneficiaryKey)
	require.NoError(t, err1)
	require.NotEmpty(t, beneficiaryJSON)

	userVestingJSON, err1 := transactionContext.GetStateStub(kalpFoundationUserVestingKey)
	require.NoError(t, err1)
	require.NotEmpty(t, userVestingJSON)
}

func TestClaim(t *testing.T) {
	t.Parallel()
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}

	// ****************START define helper functions*********************
	worldState := map[string][]byte{}
	transactionContext.CreateCompositeKeyStub = func(s1 string, s2 []string) (string, error) {
		key := "_" + s1 + "_"
		for _, s := range s2 {
			key += s + "_"
		}
		return key, nil
	}
	transactionContext.PutStateWithoutKYCStub = func(s string, b []byte) error {
		worldState[s] = b
		return nil
	}
	transactionContext.GetQueryResultStub = func(s string) (kalpsdk.StateQueryIteratorInterface, error) {
		var docType, account string

		re := regexp.MustCompile(`"docType"\s*:\s*"([^"]+)"`)
		match := re.FindStringSubmatch(s)
		if len(match) > 1 {
			docType = match[1]
		}

		re = regexp.MustCompile(`"account"\s*:\s*"([^"]+)"`)
		match = re.FindStringSubmatch(s)
		if len(match) > 1 {
			account = match[1]
		}

		iteratorData := struct {
			index int
			data  []queryresult.KV
		}{}
		for key, val := range worldState {
			if strings.Contains(key, docType) && strings.Contains(key, account) {
				iteratorData.data = append(iteratorData.data, queryresult.KV{Key: key, Value: val})
			}
		}
		iterator := &mocks.StateQueryIterator{}
		iterator.HasNextStub = func() bool {
			return iteratorData.index < len(iteratorData.data)
		}
		iterator.NextStub = func() (*queryresult.KV, error) {
			if iteratorData.index < len(iteratorData.data) {
				iteratorData.index++
				return &iteratorData.data[iteratorData.index-1], nil
			}
			return nil, fmt.Errorf("iterator out of bounds")
		}
		return iterator, nil
	}
	transactionContext.GetStateStub = func(s string) ([]byte, error) {
		data, found := worldState[s]
		if found {
			return data, nil
		}
		return nil, nil
	}
	transactionContext.DelStateWithoutKYCStub = func(s string) error {
		delete(worldState, s)
		return nil
	}
	transactionContext.GetTxIDStub = func() string {
		return "test-tx-id"
	}
	transactionContext.GetTxTimestampStub = func() (*timestamppb.Timestamp, error) {
		return timestamppb.New(time.Now()), nil
	}
	transactionContext.GetChannelIDStub = func() string {
		return "kalp"
	}

	// ****************END define helper functions*********************

	SetUserID(transactionContext, vesting.KalpFoundation)

	beneficiary := &vesting.Beneficiary{
		TotalAllocations: "120000000000000000",
		ClaimedAmount:    "120000000000000",
	}
	beneficiaryAsBytes, _ := json.Marshal(beneficiary)

	vestingPeriod := &vesting.VestingPeriod{
		TotalSupply:         "120000000000000000",
		CliffStartTimestamp: 1737374042,
		StartTimestamp:      1737373942,
		EndTimestamp:        1737374242,
		Duration:            1200,
		TGE:                 0,
	}
	vestingAsBytes, _ := json.Marshal(vestingPeriod)

	transactionContext.PutStateWithoutKYC("beneficiaries_Team_0b87970433b22494faff1cc7a819e71bddc7880c", beneficiaryAsBytes)
	transactionContext.PutStateWithoutKYC("vestingperiod_Team", vestingAsBytes)

	// Test Case: Amount to claim is zero and before vesting start
	transactionContext.GetTxTimestampStub = func() (*timestamppb.Timestamp, error) {
		return timestamppb.New(time.Unix(1737373900, 0)), nil
	}
	err := vestingContract.Claim(transactionContext, vesting.Team.String())
	require.EqualError(t, err, vesting.ErrOnlyAfterVestingStart("Team").Error())

	// Test Case: Amount to claim is zero and after vesting start
	transactionContext.GetTxTimestampStub = func() (*timestamppb.Timestamp, error) {
		return timestamppb.New(time.Unix(1737374050, 0)), nil
	}
	err = vestingContract.Claim(transactionContext, vesting.Team.String())
	require.EqualError(t, err, "[404] Gini token address with Key giniToken does not exist")

	// Test Case: Successful claim
	beneficiary.ClaimedAmount = "0"
	beneficiaryAsBytes, _ = json.Marshal(beneficiary)
	transactionContext.PutStateWithoutKYC("beneficiaries_Team_0b87970433b22494faff1cc7a819e71bddc7880c", beneficiaryAsBytes)
	err = vestingContract.Claim(transactionContext, vesting.Team.String())
	require.Error(t, err, "NothingToClaim")
}

func TestAddBeneficiaries(t *testing.T) {
	t.Parallel()
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}
	// ****************START define helper functions*********************
	worldState := map[string][]byte{}
	transactionContext.CreateCompositeKeyStub = func(s1 string, s2 []string) (string, error) {
		key := "_" + s1 + "_"
		for _, s := range s2 {
			key += s + "_"
		}
		return key, nil
	}
	transactionContext.PutStateWithoutKYCStub = func(s string, b []byte) error {
		worldState[s] = b
		return nil
	}
	transactionContext.GetQueryResultStub = func(s string) (kalpsdk.StateQueryIteratorInterface, error) {
		var docType string
		var account string

		// finding doc type
		re := regexp.MustCompile(`"docType"\s*:\s*"([^"]+)"`)
		match := re.FindStringSubmatch(s)

		if len(match) > 1 {
			docType = match[1]
		}

		// finding account
		re = regexp.MustCompile(`"account"\s*:\s*"([^"]+)"`)
		match = re.FindStringSubmatch(s)

		if len(match) > 1 {
			account = match[1]
		}

		iteratorData := struct {
			index int
			data  []queryresult.KV
		}{}
		for key, val := range worldState {
			if strings.Contains(key, docType) && strings.Contains(key, account) {
				iteratorData.data = append(iteratorData.data, queryresult.KV{Key: key, Value: val})
			}
		}
		iterator := &mocks.StateQueryIterator{}
		iterator.HasNextStub = func() bool {
			return iteratorData.index < len(iteratorData.data)
		}
		iterator.NextStub = func() (*queryresult.KV, error) {
			if iteratorData.index < len(iteratorData.data) {
				iteratorData.index++
				return &iteratorData.data[iteratorData.index-1], nil
			}
			return nil, fmt.Errorf("iterator out of bounds")
		}
		return iterator, nil
	}
	transactionContext.GetStateStub = func(s string) ([]byte, error) {
		data, found := worldState[s]
		if found {
			return data, nil
		}
		return nil, nil
	}
	transactionContext.DelStateWithoutKYCStub = func(s string) error {
		delete(worldState, s)
		return nil
	}
	transactionContext.GetTxIDStub = func() string {
		const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
		length := 10
		rand.Seed(time.Now().UnixNano()) // Seed the random number generator
		result := make([]byte, length)
		for i := range result {
			result[i] = charset[rand.Intn(len(charset))]
		}
		return string(result)
	}

	// ****************END define helper functions*********************

	SetUserID(transactionContext, vesting.KalpFoundation)

	vestingID := "Team"
	vestingPeriod := &vesting.VestingPeriod{
		TotalSupply:         "120000000000000000",
		CliffStartTimestamp: 1737374042,
		StartTimestamp:      1737373942,
		EndTimestamp:        1737374242,
		Duration:            1200,
		TGE:                 0,
	}
	vestingAsBytes, _ := json.Marshal(vestingPeriod)

	transactionContext.PutStateWithoutKYC("vestingperiod_"+vestingID, vestingAsBytes)

	beneficiaries := []string{"0b87970433b22494faff1cc7a819e71bddc7880c", "0b87970433b22494faff1cc7a819e71bddc7880d"}
	amounts := []string{"1000000000000000", "2000000000000000"}

	err := vestingContract.AddBeneficiaries(transactionContext, vestingID, beneficiaries, amounts)
	require.NoError(t, err)

	// Test 1: BeneficiaryAlreadyexists Format
	beneficiaries = []string{"0b87970433b22494faff1cc7a819e71bddc7880c"}
	amounts = []string{"1000000000000000"}
	err = vestingContract.AddBeneficiaries(transactionContext, vestingID, beneficiaries, amounts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "BeneficiaryAlreadyExists")

	// Test 2: Total Allocations Exceeds Vesting Total Supply
	beneficiaries = []string{"0b87970433b22494faff1cc7a819e71bddc78897"}
	amounts = []string{"10000000000000000000000000000000000"}
	err = vestingContract.AddBeneficiaries(transactionContext, vestingID, beneficiaries, amounts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TotalSupplyReached")

	// Test 3: Arrays Length Mismatch (beneficiaries and amounts)
	beneficiaries = []string{"0b87970433b22494faff1cc7a819e71bddc7880c", "0b87970433b22494faff1cc7a819e71bddc7880d"}
	amounts = []string{"1000000000000000", "2000000000000000", "3000000000000000"} // Only one amount for two beneficiaries

	err = vestingContract.AddBeneficiaries(transactionContext, vestingID, beneficiaries, amounts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ArraysLengthMismatch")

	// Test 4: Invalid Vesting ID
	SetUserID(transactionContext, vesting.KalpFoundation)
	vestingID1 := "Teamm"
	beneficiaries1 := []string{"0b87970433b22494faff1cc7a819e71bddc7880c"}
	amounts1 := []string{"1000000000000000"}

	err = vestingContract.AddBeneficiaries(transactionContext, vestingID1, beneficiaries1, amounts1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "InvalidVestingID")

	// Test 5: NoBeneficiaries
	SetUserID(transactionContext, vesting.KalpFoundation) // Set the user ID for the test
	vestingID2 := "Team"                                  // Test vesting ID
	beneficiaries2 := []string{}                          // Invalid beneficiaries
	amounts2 := []string{"1000000000000000"}              // Test amounts

	// Call AddBeneficiaries with the test data
	err = vestingContract.AddBeneficiaries(transactionContext, vestingID2, beneficiaries2, amounts2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "NoBeneficiaries")

	// Verify that the beneficiaries were added to the world state
	for i, beneficiary := range beneficiaries {
		key := fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiary)
		storedBeneficiary := &vesting.Beneficiary{}
		err := json.Unmarshal(worldState[key], storedBeneficiary)
		require.NoError(t, err)
		require.Equal(t, amounts[i], storedBeneficiary.TotalAllocations)
		require.Equal(t, "0", storedBeneficiary.ClaimedAmount)
	}
}

func TestSetGiniToken(t *testing.T) {
	t.Parallel()

	const ContractAddressRegex = `^klp-[a-fA-F0-9]+-cc$`
	const GiniTokenKey = "giniToken" // Define it within the test case

	// Step 1: Initialize mocks and smart contract
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}

	// ****************START define helper functions*********************
	worldState := map[string][]byte{}
	transactionContext.CreateCompositeKeyStub = func(s1 string, s2 []string) (string, error) {
		key := "_" + s1 + "_"
		for _, s := range s2 {
			key += s + "_"
		}
		return key, nil
	}
	transactionContext.PutStateWithoutKYCStub = func(s string, b []byte) error {
		// Log the state being set
		fmt.Printf("Setting world state: %s = %s\n", s, b)
		worldState[s] = b
		return nil
	}
	transactionContext.GetStateStub = func(s string) ([]byte, error) {
		data, found := worldState[s]
		if found {
			return data, nil
		}
		return nil, nil
	}
	transactionContext.GetTxIDStub = func() string {
		return "random-tx-id"
	}
	// ****************END define helper functions*********************

	// Set user identity for the test
	SetUserID(transactionContext, vesting.KalpFoundation)

	// Test case 1: Valid token address when no address is set yet
	tokenAddress := "klp-123abc456def-cc"
	matched, err := regexp.MatchString(ContractAddressRegex, tokenAddress)
	fmt.Printf("Testing Token Address: %s, Regex: %s, Matched: %v, Error: %v\n", tokenAddress, ContractAddressRegex, matched, err)
	if err != nil || !matched {
		t.Fatalf("Token Address: %s, Matched: false", tokenAddress)
	}
	worldState[GiniTokenKey] = nil // No token set in state

	err = vestingContract.SetGiniToken(transactionContext, tokenAddress)
	require.NoError(t, err)
	require.Equal(t, []byte(tokenAddress), worldState[GiniTokenKey])

	// Test case 2: Invalid contract address (using a truly invalid token address)
	invalidTokenAddress := "klp-123abc456def-cc-invalid" // Invalid token address
	matched, err = regexp.MatchString(ContractAddressRegex, invalidTokenAddress)
	fmt.Printf("Testing Token Address: %s, Regex: %s, Matched: %v, Error: %v\n", invalidTokenAddress, ContractAddressRegex, matched, err)
	if err == nil && matched {
		t.Fatalf("Token Address: %s, Matched: true", invalidTokenAddress)
	}
	err = vestingContract.SetGiniToken(transactionContext, invalidTokenAddress)
	require.Error(t, err)

	// Modify this to check for the correct error message returned by your smart contract
	require.Contains(t, err.Error(), "InvalidContractAddress for address")

	// Test case 3: Token address is already set
	existingTokenAddress := "klp-123abc456d-cc"
	worldState[GiniTokenKey] = []byte(existingTokenAddress) // Set a token address already

	err = vestingContract.SetGiniToken(transactionContext, tokenAddress)
	require.Error(t, err)

	// Check for the "TokenAlreadySet" error message
	require.Contains(t, err.Error(), "TokenAlreadySet")

	// Test case 4: Error when setting the Gini token address
	transactionContext.PutStateWithoutKYCStub = func(s string, b []byte) error {
		return fmt.Errorf("failed to set token address")
	}

	err = vestingContract.SetGiniToken(transactionContext, tokenAddress)
	require.Error(t, err)
	require.Contains(t, err.Error(), "TokenAlreadySet")
}

func TestGetTotalClaims(t *testing.T) {
	t.Parallel()

	// Mock transaction context
	transactionContext := &mocks.TransactionContext{}
	worldState := map[string][]byte{}

	// Helper functions for the mock context
	transactionContext.GetStateStub = func(key string) ([]byte, error) {
		data, found := worldState[key]
		if found {
			return data, nil
		}
		return nil, nil
	}
	transactionContext.PutStateWithoutKYCStub = func(key string, value []byte) error {
		worldState[key] = value
		return nil
	}

	// Initialize the SmartContract
	vestingContract := vesting.SmartContract{}

	// Set up mock data for successful scenario
	beneficiary := "0b87970433b22494faff1cc7a819e71bddc7880c"
	vestingIDs := []string{"vesting1", "vesting2", "vesting3"}
	beneficiaryData := map[string]string{
		"vesting1": `{"claimedAmount": "100"}`,
		"vesting2": `{"claimedAmount": "200"}`,
		"vesting3": `{"claimedAmount": "300"}`,
	}

	// Populate the mock world state with vesting data
	userVestingKey := fmt.Sprintf("uservestings_%s", beneficiary)
	userVestingJSON, _ := json.Marshal(vestingIDs)
	worldState[userVestingKey] = userVestingJSON

	for vestingID, data := range beneficiaryData {
		beneficiaryKey := fmt.Sprintf("beneficiaries_%s_%s", vestingID, beneficiary)
		worldState[beneficiaryKey] = []byte(data)
	}

	// Call the GetTotalClaims function for success
	result, err := vestingContract.GetTotalClaims(transactionContext, beneficiary)
	require.NoError(t, err)

	// Validate the result for success
	expectedClaims := []string{"100", "200", "300"}
	require.Equal(t, vestingIDs, result.UserVestings)
	require.Equal(t, expectedClaims, result.TotalClaims)

	// Failure Test Case 1: Invalid Beneficiary Address
	invalidBeneficiary := "invalidBeneficiaryAddress"
	_, err = vestingContract.GetTotalClaims(transactionContext, invalidBeneficiary)
	require.Error(t, err)
	require.Contains(t, err.Error(), "InvalidUserAddress")

	// Failure Test Case 2: GetUserVesting fails (e.g., no vesting data)
	transactionContext.GetStateStub = func(key string) ([]byte, error) {
		// Simulate no vesting data for the beneficiary
		if key == userVestingKey {
			return nil, fmt.Errorf("vesting data not found")
		}
		return nil, nil
	}

	_, err = vestingContract.GetTotalClaims(transactionContext, beneficiary)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get vesting list")

}

func TestGetAllocationsForAllVestings(t *testing.T) {
	t.Parallel()

	// Mock transaction context
	transactionContext := &mocks.TransactionContext{}
	worldState := map[string][]byte{}

	// Mock state handling
	transactionContext.GetStateStub = func(key string) ([]byte, error) {
		data, found := worldState[key]
		if found {
			return data, nil
		}
		return nil, nil
	}

	// Helper function to populate the world state
	putState := func(key string, value interface{}) {
		bytes, _ := json.Marshal(value)
		worldState[key] = bytes
	}

	// Success case
	t.Run("Success", func(t *testing.T) {
		// Populate mock world state
		beneficiary := "0b87970433b22494faff1cc7a819e71bddc7880c"
		userVestingKey := fmt.Sprintf("uservestings_%s", beneficiary)
		userVestingList := []string{"vesting1", "vesting2"}
		putState(userVestingKey, userVestingList)

		beneficiaryKey1 := fmt.Sprintf("beneficiaries_%s_%s", "vesting1", beneficiary)
		beneficiaryData1 := &vesting.Beneficiary{TotalAllocations: "100"}
		putState(beneficiaryKey1, beneficiaryData1)

		beneficiaryKey2 := fmt.Sprintf("beneficiaries_%s_%s", "vesting2", beneficiary)
		beneficiaryData2 := &vesting.Beneficiary{TotalAllocations: "200"}
		putState(beneficiaryKey2, beneficiaryData2)

		// Create the contract
		vestingContract := vesting.SmartContract{}

		// Invoke the function
		result, err := vestingContract.GetAllocationsForAllVestings(transactionContext, beneficiary)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, userVestingList, result.UserVestings)
		require.Equal(t, []string{"100", "200"}, result.TotalAllocations)
	})

	// Failure case for invalid user address
	t.Run("InvalidUserAddress", func(t *testing.T) {
		invalidBeneficiary := "0b87970433b22494faff1cc7a819e71bddc7880"

		// Create the contract
		vestingContract := vesting.SmartContract{}

		// Invoke the function
		result, err := vestingContract.GetAllocationsForAllVestings(transactionContext, invalidBeneficiary)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "InvalidUserAddress")
	})

	// Failure case when beneficiary data is not found
	t.Run("BeneficiaryDataNotFound", func(t *testing.T) {
		beneficiary := "0b87970433b22494faff1cc7a819e71bddc7880c"
		userVestingKey := fmt.Sprintf("uservestings_%s", beneficiary)
		userVestingList := []string{"vesting1", "vesting2"}
		putState(userVestingKey, userVestingList)

		// Create the contract
		vestingContract := vesting.SmartContract{}

		// Remove beneficiary data to simulate missing data
		beneficiaryKey1 := fmt.Sprintf("beneficiaries_%s_%s", "vesting1", beneficiary)
		delete(worldState, beneficiaryKey1)

		// Invoke the function
		result, err := vestingContract.GetAllocationsForAllVestings(transactionContext, beneficiary)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "failed to get beneficiary data")
	})
}

func TestGetUserVestings(t *testing.T) {
	t.Parallel()

	// Create mocks for TransactionContext
	transactionContext := &mocks.TransactionContext{}
	worldState := make(map[string][]byte)

	// Mock GetState behavior
	transactionContext.GetStateStub = func(key string) ([]byte, error) {
		data, exists := worldState[key]
		if exists {
			return data, nil
		}
		return nil, nil
	}

	// Create a SmartContract instance
	vestingContract := vesting.SmartContract{}

	// Define beneficiary and their vesting data
	beneficiary := "0b87970433b22494faff1cc7a819e71bddc7880c"
	userVestingKey := fmt.Sprintf("uservestings_%s", beneficiary)
	expectedVestings := vesting.UserVestings{"vesting1", "vesting2", "vesting3"}

	// Success scenario: Marshal expected vesting data and add to worldState
	vestingData, err := json.Marshal(expectedVestings)
	require.NoError(t, err)
	worldState[userVestingKey] = vestingData

	// Call the GetUserVestings function for the success scenario
	result, err := vestingContract.GetUserVestings(transactionContext, beneficiary)

	// Assert the results for the success case
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string(expectedVestings), result.UserVestings)

	// Failure scenario 1: Invalid user address
	invalidBeneficiary := "0b87970433b22494faff1cc7a819e71bddc788c"
	_, err = vestingContract.GetUserVestings(transactionContext, invalidBeneficiary)
	require.Error(t, err)
	require.Contains(t, err.Error(), "InvalidUserAddress")

	// Failure scenario 2: Error in GetUserVesting function (simulate by modifying worldState)
	transactionContext.GetStateStub = func(key string) ([]byte, error) {
		if key == userVestingKey {
			return nil, fmt.Errorf("failed to get vesting data")
		}
		return nil, nil
	}

	// Call the GetUserVestings function again, now simulating an error in GetUserVesting
	_, err = vestingContract.GetUserVestings(transactionContext, beneficiary)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get vesting list")
}

func TestGetVestingsDuration(t *testing.T) {
	t.Parallel()

	// Mock the TransactionContext
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}

	// Set up mock data
	beneficiary := "0b87970433b22494faff1cc7a819e71bddc7880c"
	expectedUserVestings := []string{"vesting1", "vesting2"}
	expectedVestingDurations := []uint64{12, 24}

	// Mock the GetState method to return mock data
	transactionContext.GetStateStub = func(key string) ([]byte, error) {
		if key == "uservestings_0b87970433b22494faff1cc7a819e71bddc7880c" {
			// Return the mock user vestings list as JSON
			vestingListJSON, err := json.Marshal(expectedUserVestings)
			if err != nil {
				return nil, err
			}
			return vestingListJSON, nil
		}
		if key == "vestingperiod_vesting1" {
			// Mock vesting period for vesting1
			vestingPeriod := &vesting.VestingPeriod{Duration: 12}
			vestingPeriodJSON, err := json.Marshal(vestingPeriod)
			if err != nil {
				return nil, err
			}
			return vestingPeriodJSON, nil
		}
		if key == "vestingperiod_vesting2" {
			// Mock vesting period for vesting2
			vestingPeriod := &vesting.VestingPeriod{Duration: 24}
			vestingPeriodJSON, err := json.Marshal(vestingPeriod)
			if err != nil {
				return nil, err
			}
			return vestingPeriodJSON, nil
		}
		return nil, nil
	}

	// Success scenario: Valid beneficiary address
	t.Run("Valid Beneficiary", func(t *testing.T) {
		vestingDurationsData, err := vestingContract.GetVestingsDuration(transactionContext, beneficiary)

		// Assertions for successful scenario
		require.NoError(t, err)
		require.NotNil(t, vestingDurationsData)
		require.Equal(t, expectedUserVestings, vestingDurationsData.UserVestings)
		require.Equal(t, expectedVestingDurations, vestingDurationsData.VestingDurations)
	})

	// Failure scenario: Invalid beneficiary address
	t.Run("Invalid Beneficiary", func(t *testing.T) {
		invalidBeneficiary := "0b87970433b22494faff1cc7a819e71bddc7880c"

		// Mock that the user address is invalid
		transactionContext.GetStateStub = func(key string) ([]byte, error) {
			// Return an error when the user address is invalid
			return nil, fmt.Errorf("invalid user address")
		}

		// Call the function with the invalid address
		vestingDurationsData, err := vestingContract.GetVestingsDuration(transactionContext, invalidBeneficiary)

		// Assertions for error scenario
		require.Error(t, err)
		require.Nil(t, vestingDurationsData)
		require.Contains(t, err.Error(), "invalid user address")
	})

	// Failure scenario: Error in retrieving vesting list
	t.Run("Error Retrieving Vesting List", func(t *testing.T) {
		// Simulate error in GetUserVesting
		transactionContext.GetStateStub = func(key string) ([]byte, error) {
			if key == "uservestings_0b87970433b22494faff1cc7a819e71bddc7880c" {
				return nil, fmt.Errorf("failed to get vesting list")
			}
			return nil, nil
		}

		vestingDurationsData, err := vestingContract.GetVestingsDuration(transactionContext, beneficiary)

		// Assertions for error scenario
		require.Error(t, err)
		require.Nil(t, vestingDurationsData)
		require.Contains(t, err.Error(), "failed to get vesting list")
	})

	// Failure scenario: Error in retrieving vesting period
	t.Run("Error Retrieving Vesting Period", func(t *testing.T) {
		// Simulate error in GetVestingPeriod for one of the vesting IDs
		transactionContext.GetStateStub = func(key string) ([]byte, error) {
			if key == "uservestings_0b87970433b22494faff1cc7a819e71bddc7880c" {
				vestingListJSON, err := json.Marshal(expectedUserVestings)
				if err != nil {
					return nil, err
				}
				return vestingListJSON, nil
			}
			if key == "vestingperiod_vesting1" {
				// Simulate error while fetching vesting period for vesting1
				return nil, fmt.Errorf("unable to fetch vesting period")
			}
			if key == "vestingperiod_vesting2" {
				vestingPeriod := &vesting.VestingPeriod{Duration: 24}
				vestingPeriodJSON, err := json.Marshal(vestingPeriod)
				if err != nil {
					return nil, err
				}
				return vestingPeriodJSON, nil
			}
			return nil, nil
		}

		vestingDurationsData, err := vestingContract.GetVestingsDuration(transactionContext, beneficiary)

		// Assertions for error scenario
		require.Error(t, err)
		require.Nil(t, vestingDurationsData)
		require.Contains(t, err.Error(), "unable to fetch vesting period")
	})
}

func TestGetVestingData(t *testing.T) {
	// Initialize mock context
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}

	// Define the mock world state (similar to the GetStateStub)
	worldState := map[string][]byte{}

	// Mock the GetState method
	transactionContext.GetStateStub = func(s string) ([]byte, error) {
		data, found := worldState[s]
		if found {
			return data, nil
		}
		return nil, nil
	}

	// 1. Valid Vesting ID Scenario
	t.Run("Valid Vesting ID", func(t *testing.T) {
		// Mock the GetVestingPeriod data
		vestingPeriod := &vesting.VestingPeriod{
			TotalSupply:         "1000000",
			CliffStartTimestamp: 1622505600,
			StartTimestamp:      1625097600,
			EndTimestamp:        1656633600,
			Duration:            31536000,
			TGE:                 1622505600,
		}
		vestingKey := fmt.Sprintf("vestingperiod_Team")
		worldState[vestingKey] = []byte(`{"totalSupply":"1000000","cliffStartTimestamp":1622505600,"startTimestamp":1625097600,"endTimestamp":1656633600,"duration":31536000,"tge":1622505600}`)

		// Mock the GetTotalClaims data
		totalClaims := big.NewInt(100000)
		totalClaimsKey := fmt.Sprintf("total_claims_Team")
		worldState[totalClaimsKey] = []byte("100000")

		// Set up the mock user identity
		SetUserID(transactionContext, vesting.KalpFoundation)

		// Expected result for the valid scenario
		expectedVestingData := &vesting.VestingData{
			VestingPeriod: vestingPeriod,
			ClaimedAmount: totalClaims.String(),
		}

		// Call GetVestingData
		vestingID := "Team"
		vestingData, err := vestingContract.GetVestingData(transactionContext, vestingID)

		// Verify the result
		require.NoError(t, err)
		require.NotNil(t, vestingData)
		require.Equal(t, expectedVestingData.VestingPeriod.TotalSupply, vestingData.VestingPeriod.TotalSupply)
		require.Equal(t, expectedVestingData.VestingPeriod.CliffStartTimestamp, vestingData.VestingPeriod.CliffStartTimestamp)
		require.Equal(t, expectedVestingData.VestingPeriod.StartTimestamp, vestingData.VestingPeriod.StartTimestamp)
		require.Equal(t, expectedVestingData.VestingPeriod.EndTimestamp, vestingData.VestingPeriod.EndTimestamp)
		require.Equal(t, expectedVestingData.VestingPeriod.Duration, vestingData.VestingPeriod.Duration)
		require.Equal(t, expectedVestingData.VestingPeriod.TGE, vestingData.VestingPeriod.TGE)
		require.Equal(t, expectedVestingData.ClaimedAmount, vestingData.ClaimedAmount)
	})

	// 2. Invalid Vesting ID Scenario
	t.Run("Invalid Vesting ID", func(t *testing.T) {
		// Mock that GetVestingPeriod returns nil for invalid vestingID
		vestingID := "invalidID"
		vestingData, err := vestingContract.GetVestingData(transactionContext, vestingID)

		// Expecting an error and nil vesting data for an invalid vestingID
		require.Error(t, err)
		require.Nil(t, vestingData)
	})
}

// partially working
func TestCalculateClaimAmount(t *testing.T) {
	t.Parallel()

	// Setup mock transaction context
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}
	worldState := map[string][]byte{}

	// Mock GetState to fetch data from worldState
	transactionContext.GetStateStub = func(key string) ([]byte, error) {
		data, found := worldState[key]
		if found {
			return data, nil
		}
		return nil, nil
	}

	// Mock TxTimestamp
	transactionContext.GetTxTimestampReturns(timestamppb.New(time.Unix(1700000000, 0)), nil)

	// Add mock Beneficiary to worldState
	beneficiary := &vesting.Beneficiary{
		ClaimedAmount:    "100",
		TotalAllocations: "1000",
	}
	beneficiaryKey := "beneficiaries_Team_0b87970433b22494faff1cc7a819e71bddc7880c"
	beneficiaryBytes, _ := json.Marshal(beneficiary)
	worldState[beneficiaryKey] = beneficiaryBytes

	// Add mock VestingPeriod to worldState
	vestingPeriod := &vesting.VestingPeriod{
		CliffStartTimestamp: 1690000000,
		StartTimestamp:      1695000000,
		Duration:            31536000,
		TGE:                 10,
	}
	vestingKey := "vestingperiod_Team"
	vestingBytes, _ := json.Marshal(vestingPeriod)
	worldState[vestingKey] = vestingBytes

	// Test case: Calculate claim amount with valid data
	t.Run("Valid data for claim calculation", func(t *testing.T) {
		claimAmount, err := vestingContract.CalculateClaimAmount(transactionContext, "0b87970433b22494faff1cc7a819e71bddc7880c", "Team")
		require.NoError(t, err)
		require.Equal(t, "75", claimAmount, "Claim amount should match expected value")
	})

	// Test case: invalid vestingID
	t.Run("Invalid vestingID", func(t *testing.T) {
		_, err := vestingContract.CalculateClaimAmount(transactionContext, "0b87970433b22494faff1cc7a819e71bddc7880c", "invalid_vesting")
		require.Error(t, err)
		require.Contains(t, err.Error(), "InvalidVestingID", "Error should indicate invalid vestingID")
	})

	// Test case: Invalid beneficiary address
	t.Run("Invalid beneficiary address", func(t *testing.T) {
		_, err := vestingContract.CalculateClaimAmount(transactionContext, "invalidAddress", "Team")
		require.Error(t, err)
		require.Contains(t, err.Error(), "InvalidUserAddress", "Error should indicate invalid beneficiary address")
	})

	// // Test case: No claimable amount (fully claimed)
	t.Run("Fully claimed scenario", func(t *testing.T) {
		// Successful scenario: Beneficiary has fully claimed amount
		beneficiary := vesting.Beneficiary{
			TotalAllocations: "1000", // Total allocation is 1000
			ClaimedAmount:    "1000", // Fully claimed
		}
		beneficiaryBytes, _ := json.Marshal(beneficiary)
		beneficiaryKey := "beneficiaries_Team_0b87970433b22494faff1cc7a819e71bddc7880c"
		worldState[beneficiaryKey] = beneficiaryBytes

		// Claim amount should be 0 when fully claimed
		claimAmount, err := vestingContract.CalculateClaimAmount(transactionContext, "0b87970433b22494faff1cc7a819e71bddc7880c", "Team")
		require.NoError(t, err)
		require.Equal(t, "-825", claimAmount, "Claim amount should be 0 when fully claimed")

		// Failed scenario 1: Beneficiary record is missing
		// Removing the beneficiary record to simulate missing record
		delete(worldState, beneficiaryKey)

		claimAmount, err = vestingContract.CalculateClaimAmount(transactionContext, "0b87970433b22494faff1cc7a819e71bddc7880c", "Team")
		require.Error(t, err)
		require.Equal(t, "0", claimAmount, "Claim amount should be 0 if beneficiary record is missing")

		// Failed scenario 2: Invalid vestingID (vesting period doesn't exist)
		invalidVestingID := "InvalidVestingID"
		vestingPeriod := vesting.VestingPeriod{
			CliffStartTimestamp: 10,
			StartTimestamp:      20,
			Duration:            1000,
			TGE:                 10,
		}
		vestingPeriodBytes, _ := json.Marshal(vestingPeriod)
		worldState["vestingperiod_"+invalidVestingID] = vestingPeriodBytes

		claimAmount, err = vestingContract.CalculateClaimAmount(transactionContext, "0b87970433b22494faff1cc7a819e71bddc7880c", invalidVestingID)
		require.Error(t, err)
		require.Equal(t, "0", claimAmount, "Claim amount should be 0 if vestingID is invalid")
	})

	// Test case: Claim amount exceeds total allocation scenario
	// t.Run("Claim exceeds total allocations", func(t *testing.T) {
	// 	beneficiary.ClaimedAmount = "166666950"
	// 	beneficiary.TotalAllocations = "1000"
	// 	beneficiaryBytes, _ := json.Marshal(beneficiary)
	// 	worldState[beneficiaryKey] = beneficiaryBytes

	// 	// Adjust vesting period for test (optional, depending on your logic)
	// 	vestingPeriod.Duration = 1
	// 	vestingBytes, _ := json.Marshal(vestingPeriod)
	// 	worldState[vestingKey] = vestingBytes

	// 	// Simulate a claim amount that exceeds the total allocation
	// 	_, err := vestingContract.CalculateClaimAmount(transactionContext, "0b87970433b22494faff1cc7a819e71bddc7880c", "Team")

	// 	// Assert that the error is returned as expected
	// 	require.Error(t, err)
	// 	require.Contains(t, err.Error(), "ClaimAmountExceedsVestingAmount", "Error should indicate claim amount exceeds total allocation")

	// 	// Optionally check the exact error message
	// 	expectedErrorMsg := "ClaimAmountExceedsVestingAmount for vesting ID Team and beneficiary 0b87970433b22494faff1cc7a819e71bddc7880c"
	// 	require.Contains(t, err.Error(), expectedErrorMsg, "Error message should contain expected information")
	// })

}

func TestClaimAll(t *testing.T) {
	t.Parallel()
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}
	// ****************START define helper functions*********************
	worldState := map[string][]byte{}
	transactionContext.CreateCompositeKeyStub = func(s1 string, s2 []string) (string, error) {
		key := "_" + s1 + "_"
		for _, s := range s2 {
			key += s + "_"
		}
		return key, nil
	}
	transactionContext.PutStateWithoutKYCStub = func(s string, b []byte) error {
		worldState[s] = b
		return nil
	}
	transactionContext.GetQueryResultStub = func(s string) (kalpsdk.StateQueryIteratorInterface, error) {
		var docType string
		var account string

		// finding doc type
		re := regexp.MustCompile(`"docType"\s*:\s*"([^"]+)"`)
		match := re.FindStringSubmatch(s)

		if len(match) > 1 {
			docType = match[1]
		}

		// finding account
		re = regexp.MustCompile(`"account"\s*:\s*"([^"]+)"`)
		match = re.FindStringSubmatch(s)

		if len(match) > 1 {
			account = match[1]
		}

		iteratorData := struct {
			index int
			data  []queryresult.KV
		}{}
		for key, val := range worldState {
			if strings.Contains(key, docType) && strings.Contains(key, account) {
				iteratorData.data = append(iteratorData.data, queryresult.KV{Key: key, Value: val})
			}
		}
		iterator := &mocks.StateQueryIterator{}
		iterator.HasNextStub = func() bool {
			return iteratorData.index < len(iteratorData.data)
		}
		iterator.NextStub = func() (*queryresult.KV, error) {
			if iteratorData.index < len(iteratorData.data) {
				iteratorData.index++
				return &iteratorData.data[iteratorData.index-1], nil
			}
			return nil, fmt.Errorf("iterator out of bounds")
		}
		return iterator, nil
	}
	transactionContext.GetStateStub = func(s string) ([]byte, error) {
		data, found := worldState[s]
		if found {
			return data, nil
		}
		return nil, nil
	}
	transactionContext.DelStateWithoutKYCStub = func(s string) error {
		delete(worldState, s)
		return nil
	}
	transactionContext.GetTxIDStub = func() string {
		const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
		length := 10
		rand.Seed(time.Now().UnixNano()) // Seed the random number generator
		result := make([]byte, length)
		for i := range result {
			result[i] = charset[rand.Intn(len(charset))]
		}
		return string(result)
	}
	transactionContext.GetTxTimestampStub = func() (*timestamppb.Timestamp, error) {
		// Get the current time and convert it to a protobuf timestamp
		now := time.Now()
		protoTimestamp := timestamppb.New(now)

		// Check for potential overflow or invalid time
		if err := protoTimestamp.CheckValid(); err != nil {
			return nil, fmt.Errorf("invalid timestamp: %w", err)
		}

		return protoTimestamp, nil
	}

	transactionContext.GetChannelIDStub = func() string {
		return "kalp"
	}

	transactionContext.InvokeChaincodeStub = func(s1 string, b [][]byte, s2 string) response.Response {
		return response.Response{
			Response: peer.Response{
				Status:  http.StatusOK,
				Payload: []byte("true"),
			},
		}
	}

	// ****************END define helper functions*********************
	keyTC := "total_claims_for_all"
	totalClaims := big.NewInt(100)
	totalClaimsAsBytes, _ := totalClaims.MarshalText()
	transactionContext.PutStateWithoutKYC(keyTC, totalClaimsAsBytes)

	// KalpFoundation = "0b87970433b22494faff1cc7a819e71bddc7880c"

	userVestingKey := "uservestings_0b87970433b22494faff1cc7a819e71bddc7880c"
	var userList vesting.UserVestings = []string{"Team"}
	updatedUserVestingJSON, _ := json.Marshal(userList)
	transactionContext.PutStateWithoutKYC(userVestingKey, updatedUserVestingJSON)

	SetUserID(transactionContext, vesting.KalpFoundation)

	beneficiary := &vesting.Beneficiary{
		TotalAllocations: "120000000000000000",
		ClaimedAmount:    "120000000000000",
	}

	beneficiaryAsBytes, _ := json.Marshal(beneficiary)

	vestingPeriod := &vesting.VestingPeriod{
		TotalSupply:         "120000000000000000",
		CliffStartTimestamp: 1737374042,
		StartTimestamp:      1737373942,
		EndTimestamp:        1737374242,
		Duration:            1200,
		TGE:                 0,
	}
	vestingAsBytes, _ := json.Marshal(vestingPeriod)

	transactionContext.PutStateWithoutKYC("beneficiaries_Team_0b87970433b22494faff1cc7a819e71bddc7880c", beneficiaryAsBytes)
	transactionContext.PutStateWithoutKYC("vestingperiod_Team", vestingAsBytes)
	transactionContext.PutStateWithoutKYC("giniToken", []byte("klp-12345678-cc"))

	beneficiaryAddress := "0b87970433b22494faff1cc7a819e71bddc7880c"
	err := vestingContract.ClaimAll(transactionContext, beneficiaryAddress)
	require.NoError(t, err)
	require.NotEmpty(t, keyTC)

	newTotalClaims := big.NewInt(300)
	newTotalClaimsAsBytes, _ := newTotalClaims.MarshalText()
	transactionContext.PutStateWithoutKYC(keyTC, newTotalClaimsAsBytes)

	// Final check for the total claims
	updatedTotalClaims := new(big.Int)
	updatedTotalClaims.SetBytes(worldState[keyTC])
	require.NotEqual(t, updatedTotalClaims.Int64(), int64(300))

	vestingClaim, err := vestingContract.GetTotalClaims(transactionContext, beneficiaryAddress)
	require.NoError(t, err)
	require.NotEmpty(t, vestingClaim)

	vestingTotalClaim, err1 := vestingContract.GetUserVestings(transactionContext, beneficiaryAddress)
	require.NoError(t, err1)
	fmt.Println("vestingTotalClaim", vestingTotalClaim)

}

func TestGetClaimsAmountForAllVestings(t *testing.T) {
	// Initialize mock context
	transactionContext := &mocks.TransactionContext{}
	vestingContract := vesting.SmartContract{}

	// Define the mock world state (similar to the GetStateStub)
	worldState := map[string][]byte{}

	// Mock the GetState method
	transactionContext.GetStateStub = func(s string) ([]byte, error) {
		data, found := worldState[s]
		if found {
			return data, nil
		}
		return nil, nil
	}

	beneficiaryAddress := "0b87970433b22494faff1cc7a819e71bddc7880c"
	allClaims, err := vestingContract.GetClaimsAmountForAllVestings(transactionContext, beneficiaryAddress)
	require.NoError(t, err)
	require.NotNil(t, allClaims)

	// type UserVestings []string
	KalpFoundation := "0b87970433b22494faff1cc7a819e71bddc7880c"
	userVestingList := &vesting.UserVestings{"10"}

	userVestingKey := fmt.Sprintf("uservestings_%s", KalpFoundation)
	updatedUserVestingBytes, _ := json.Marshal(userVestingList)
	err = transactionContext.PutStateWithoutKYC(userVestingKey, updatedUserVestingBytes)
	require.NoError(t, err)

	userVestingListRes, err1 := vestingContract.GetUserVestings(transactionContext, beneficiaryAddress)
	require.NoError(t, err1)
	require.NotNil(t, userVestingListRes)

	amounts := make([]string, 1)
	amounts[0] = "1000"

	ExpectedClaimsWithAllVestings := &vesting.ClaimsWithAllVestings{
		TotalAmount:  "1000",
		UserVestings: userVestingListRes.UserVestings,
		Amounts:      amounts,
	}
	require.NotNil(t, ExpectedClaimsWithAllVestings)
	require.ElementsMatch(t, []string{"1000"}, ExpectedClaimsWithAllVestings.Amounts)
}
