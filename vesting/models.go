package vesting

type Beneficiary struct {
	TotalAllocations string
	ClaimedAmount    string
}

type VestingPeriod struct {
	TotalSupply         string
	CliffStartTimestamp uint64
	StartTimestamp      uint64
	EndTimestamp        uint64
	Duration            uint64
	TGE                 uint64
}

type VestingPeriodEvent struct {
	VestingID           string
	TotalSupply         string
	CliffStartTimestamp uint64
	StartTimestamp      uint64
	EndTimestamp        uint64
	TGE                 uint64
}
