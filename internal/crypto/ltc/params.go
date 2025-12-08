package ltc

import (
	"math/big"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
)

// LitecoinParams defines the network parameters for the Litecoin main network.
var LitecoinParams = chaincfg.Params{
	Name:        "litecoin",
	Net:         wire.BitcoinNet(0xdbb6c0fb), // MainNet Magic
	DefaultPort: "9333",
	DNSSeeds:    []chaincfg.DNSSeed{}, // Not needed for RPC
	
	// Chain parameters
	GenesisHash:      nil, // Not needed for this usage
	PowLimit:         new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1)),
	PowLimitBits:     0x1e0ffff0,
	BIP0034Height:    710000,
	BIP0065Height:    918684,
	BIP0066Height:    811879,
	CoinbaseMaturity: 100,
	SubsidyReductionInterval: 840000,
	TargetTimespan:           ((3 * 24 + 12) * 60 * 60), // 3.5 days
	TargetTimePerBlock:       (2 * 60) + 30,             // 2.5 mins
	RetargetAdjustmentFactor: 4,                         // 25% less, 400% more
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0,
	GenerateSupported:        false,

	// Checkpoints
	Checkpoints: nil,

	// Enforce current block height matches script rule change
	// RuleChangeActivationThreshold: 6048, // 75% of 8064
	// MinerConfirmationWindow:       8064, // 2016 * 4

	// Address encoding magics
	PubKeyHashAddrID: 0x30, // 48 (L)
	ScriptHashAddrID: 0x32, // 50 (M)
	PrivateKeyID:     0xb0, // 176

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // xprv
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // xpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 2,

	// Bech32 HRP
	Bech32HRPSegwit: "ltc",
}
