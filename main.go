package main

import (
	"log"

	"github.com/p2eengineering/gini-vesting-contract/vesting"
	"github.com/p2eengineering/kalp-sdk-public/kalpsdk"
)

func main() {
	contract := kalpsdk.Contract{}
	contract.Logger = kalpsdk.NewLogger()
	vestingChaincode, err := kalpsdk.NewChaincode(&vesting.SmartContract{Contract: contract})
	if err != nil {
		log.Panicf("Error creating vesting chaincode: %v", err)
	}

	if err := vestingChaincode.Start(); err != nil {
		log.Panicf("Error starting vesting chaincode: %v", err)
	}
}
