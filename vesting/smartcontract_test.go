package vesting_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

	// transactionContext.GetStateStub(vestingAsBytes, nil)
	// transactionContext.GetStateReturns(beneficiaryAsBytes, nil)
	err := vestingContract.Claim(transactionContext, vesting.Team.String())
	require.NoError(t, err)
}
