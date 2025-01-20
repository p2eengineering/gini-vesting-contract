package vesting

type TokenAllocation int

const (
	KalpFoundation                     = "0b87970433b22494faff1cc7a819e71bddc7880c"
	KalpFoundationTotalAllocations     = "560000000000000000000000000"
	KalpFoundationClaimedAmount        = "11200000000000000000000000"
	KalpFoundationBeneficiaryKeyPrefix = "beneficiaries_EcosystemReserve_"
	KalpFoundationUserVestingKeyPrefix = "uservestings_"
	ContractAddressRegex               = `^klp-[a-fA-F0-9]+-cc$`
	HexAddressRegex                    = `^[0-9a-fA-F]{40}$`
	GiniTokenEvent                     = "SetGiniToken"
	KalpFoundationKey                  = "kalp_foundation"
	ClaimInterval                      = 30 * 24 * 60 * 60

	GiniTransfer = "Transfer"
	GiniTokenKey = "giniToken"

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
