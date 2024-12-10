package vesting

type TokenAllocation int

const (
	kalpFoundation                 = "0b87970433b22494faff1cc7a819e71bddc7880c"
	kalpFoundationTotalAllocations = "560000000000000000000000000"
	kalpFoundationClaimedAmount    = "11200000000000000000000000"
	kalpFoundationBeneficiaryKey   = "beneficiaries_EcosystemReserve_kalp_foundation"
	kalpFoundationUserVestingKey   = "uservestings_kalp_foundation"
	contractAddressRegex           = `^klp.*cc$`
	hexAddressRegex                = `^0x[0-9a-fA-F]{40}$`
)

const (
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
