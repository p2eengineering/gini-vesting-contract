package vesting

type TokenAllocation int

const (
	kalpFoundation = "0b87970433b22494faff1cc7a819e71bddc7880c"
	// kalpFoundation                 = "user1"
	kalpFoundationTotalAllocations = "560000000000000000000000000"
	kalpFoundationClaimedAmount    = "11200000000000000000000000"
	kalpFoundationBeneficiary      = "beneficiaries_EcosystemReserve_kalp_foundation"
	kalpFoundationUserVesting      = "uservesting_kalp_foundation"
	claimInterval                  = 30
	giniTokenEvent                 = "SetGiniToken"

	Team TokenAllocation = iota
	Foundation
	AngelRound
	SeedRound
	PrivateRound1
	PrivateRound2
	Advisors
	KOLRound
	Marketing
	StakingRewards
	EcosystemReserve
	Airdrop
	LiquidityPool
	PublicAllocation
)

func (t TokenAllocation) String() string {
	return [...]string{
		"Team",
		"Foundation",
		"AngelRound",
		"SeedRound",
		"PrivateRound1",
		"PrivateRound2",
		"Advisors",
		"KOLRound",
		"Marketing",
		"StakingRewards",
		"EcosystemReserve",
		"Airdrop",
		"LiquidityPool",
		"PublicAllocation",
	}[t]
}
