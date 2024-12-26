package vesting

type TokenAllocation int

const (
	kalpFoundation                     = "0b87970433b22494faff1cc7a819e71bddc7880c"
	kalpFoundationTotalAllocations     = "560000000000000000000000000"
	kalpFoundationClaimedAmount        = "11200000000000000000000000"
	kalpFoundationBeneficiaryKeyPrefix = "beneficiaries_EcosystemReserve_"
	kalpFoundationUserVestingKeyPrefix = "uservestings_"
	contractAddressRegex               = `^klp-[a-fA-F0-9]+-cc$`
	hexAddressRegex                    = `^[0-9a-fA-F]{40}$`
	giniTokenEvent                     = "SetGiniToken"
	kalpFoundationKey                  = "kalp_foundation"
	claimInterval                      = 30 * 24 * 60 * 60

	giniTransfer = "Transfer"
	giniTokenKey = "giniToken"

	// Events Keys
	ClaimKey              = "Claim"
	BeneficiariesAddedKey = "BeneficiariesAdded"
	VestingInitializedKey = "VestingInitialized"
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
