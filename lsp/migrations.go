package lsp

import "github.com/Sora233/DDBOT/v2/lsp/version"

const LspVersionName = "lsp"
const LspSupportVersion int64 = 1

// 对于DDBOT来说，Version name固定为lsp，升级后的版本无法回退

var lspMigrationMap = version.NewMigrationMapFromMap(
	map[int64]version.Migration{
		0: new(V1),
	},
)
