package config

import (
	"github.com/ethereum/go-ethereum/params"
)

var (
	Bootnodes = map[string][]string{
		"main": {
			"enode://5d1a3b86c8ee297142fb916db77785cc679d4ed10f51cca84c9ee9442d66280a90378d9ca364620bd8cebec0aafb2cf22cd033905afe2028c22b804a2e18d981@bootnode0.x1-testnet.xen.network:5050",
			"enode://703e4f01760a1f432d5e5e9a8a9c86b5dca301563bc9559f8c991de052e438d8072b2e789f7722762ea44c5c06ed87428a25606f278a128d7a39bd5996be9536@bootnode1.x1-testnet.xen.network:5050",
			"enode://873de92942769e57c2360dd3198af2fda9602a9826b1f1a6d36326b6a75c9178c4f13f910c9e715a5ca6344aa62452930fcee18636ab0a95f269dded52f45960@bootnode2.x1-testnet.xen.network:5050",
			"enode://125e59370d9c7f5a195e2b99da9bb4172c8ac7dde74bb47a5b97fdfbeb3e876ee64289fd05b96b2b24006e00c4a6ad990ae159008f006f809966636797ab0054@bootnode3.x1-testnet.xen.network:5050",
		},
		"test": {
			"enode://5d1a3b86c8ee297142fb916db77785cc679d4ed10f51cca84c9ee9442d66280a90378d9ca364620bd8cebec0aafb2cf22cd033905afe2028c22b804a2e18d981@bootnode0.x1-testnet.xen.network:5050",
			"enode://703e4f01760a1f432d5e5e9a8a9c86b5dca301563bc9559f8c991de052e438d8072b2e789f7722762ea44c5c06ed87428a25606f278a128d7a39bd5996be9536@bootnode1.x1-testnet.xen.network:5050",
			"enode://873de92942769e57c2360dd3198af2fda9602a9826b1f1a6d36326b6a75c9178c4f13f910c9e715a5ca6344aa62452930fcee18636ab0a95f269dded52f45960@bootnode2.x1-testnet.xen.network:5050",
			"enode://125e59370d9c7f5a195e2b99da9bb4172c8ac7dde74bb47a5b97fdfbeb3e876ee64289fd05b96b2b24006e00c4a6ad990ae159008f006f809966636797ab0054@bootnode3.x1-testnet.xen.network:5050",
		},
	}
)

func OverrideParams() {
	params.MainnetBootnodes = []string{}
	params.RopstenBootnodes = []string{}
	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
