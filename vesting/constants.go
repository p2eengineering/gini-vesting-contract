package vesting

type TokenAllocation int

const (
	kalpFoundation                                 = "0b87970433b22494faff1cc7a819e71bddc7880c"
	kalpFoundationTotalAllocations                 = "560000000000000000000000000"
	kalpFoundationClaimedAmount                    = "11200000000000000000000000"
	contractAddressRegex                           = `^klp-[a-fA-F0-9]+-cc$`
	hexAddressRegex                                = `^[0-9a-fA-F]{40}$`
	giniTokenEvent                                 = "SetGiniToken"
	claimInterval                                  = 30 * 24 * 60 * 60
	EcosystemReserveTotalSupplyAfterInitialisation = 0

	giniTransfer = "Transfer"
	giniTokenKey = "giniToken"

	// Events Keys
	ClaimKey                              = "Claim"
	BeneficiariesAddedKey                 = "BeneficiariesAdded"
	VestingInitializedKey                 = "VestingInitialized"
	EcoSystemReserveTotalSupplyChangedKey = "EcoSystemReserveTotalSupplyChanged"

	// Durations
	// Team Vesting Configuration
	TeamCliffDuration   = 30 * 12 * 24 * 60 * 60
	TeamVestingDuration = 30 * 24 * 24 * 60 * 60
	TeamTotalSupply     = 300000000
	TeamTGE             = 0

	// Foundation Vesting Configuration
	FoundationCliffDuration   = 0
	FoundationVestingDuration = 30 * 12 * 24 * 60 * 60
	FoundationTotalSupply     = 220000000
	FoundationTGE             = 0

	// Private Round 1 Vesting Configuration
	PrivateRound1CliffDuration   = 30 * 12 * 24 * 60 * 60
	PrivateRound1VestingDuration = 30 * 12 * 24 * 60 * 60
	PrivateRound1TotalSupply     = 200000000
	PrivateRound1TGE             = 0

	// Private Round 2 Vesting Configuration
	PrivateRound2CliffDuration   = 30 * 6 * 24 * 60 * 60
	PrivateRound2VestingDuration = 30 * 12 * 24 * 60 * 60
	PrivateRound2TotalSupply     = 60000000
	PrivateRound2TGE             = 0

	// Advisors Vesting Configuration
	AdvisorsCliffDuration   = 30 * 9 * 24 * 60 * 60
	AdvisorsVestingDuration = 30 * 12 * 24 * 60 * 60
	AdvisorsTotalSupply     = 30000000
	AdvisorsTGE             = 0

	// KOL Round Vesting Configuration
	KOLRoundCliffDuration   = 30 * 3 * 24 * 60 * 60
	KOLRoundVestingDuration = 30 * 6 * 24 * 60 * 60
	KOLRoundTotalSupply     = 30000000
	KOLRoundTGE             = 25

	// Marketing Vesting Configuration
	MarketingCliffDuration   = 30 * 1 * 24 * 60 * 60
	MarketingVestingDuration = 30 * 18 * 24 * 60 * 60
	MarketingTotalSupply     = 80000000
	MarketingTGE             = 10

	// Staking Rewards Vesting Configuration
	StakingRewardsCliffDuration   = 30 * 3 * 24 * 60 * 60
	StakingRewardsVestingDuration = 30 * 24 * 24 * 60 * 60
	StakingRewardsTotalSupply     = 180000000
	StakingRewardsTGE             = 0

	// Ecosystem Reserve Vesting Configuration
	EcosystemReserveCliffDuration   = 0
	EcosystemReserveVestingDuration = 30 * 150 * 24 * 60 * 60
	EcosystemReserveTotalSupply     = 560000000
	EcosystemReserveTGE             = 2

	// Airdrop Vesting Configuration
	AirdropCliffDuration   = 30 * 6 * 24 * 60 * 60
	AirdropVestingDuration = 30 * 9 * 24 * 60 * 60
	AirdropTotalSupply     = 80000000
	AirdropTGE             = 10

	// Liquidity Pool Vesting Configuration
	LiquidityPoolCliffDuration   = 0
	LiquidityPoolVestingDuration = 30 * 6 * 24 * 60 * 60
	LiquidityPoolTotalSupply     = 200000000
	LiquidityPoolTGE             = 25

	// Public Allocation Vesting Configuration
	PublicAllocationCliffDuration   = 30 * 3 * 24 * 60 * 60
	PublicAllocationVestingDuration = 30 * 6 * 24 * 60 * 60
	PublicAllocationTotalSupply     = 60000000
	PublicAllocationTGE             = 25

	// Composite Keys
	BeneficiariesPrefix  = "beneficiaries"
	VestingPeriodPrefix  = "vestingperiod"
	UserVestingsPrefix   = "uservestings"
	TotalClaimsForAllKey = "total_claims_for_all"
	TotalClaimsPrefix    = "total_claims"
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
