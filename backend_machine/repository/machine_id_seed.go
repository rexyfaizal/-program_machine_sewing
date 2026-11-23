package repository

import "strings"

// Seed awal Machine ID dari mapping IE (UUID → Machine ID).
// Hanya diisi jika machine_id di DB masih kosong.
var machineIDSeedByUUID = map[string]string{
	"216fccp6vgdrkxct":         "JUKI-0001",
	"21icjlfiheeph5d":          "JUKI-0002",
	"21c6mza3judbflhh":         "JUKI-0003",
	"218s9piz8x9ajg2l":         "JUKI-0004",
	"21bborkkws72ff5a":         "JUKI-0005",
	"214ahhyuxnbpnayv":         "JUKI-0006",
	"214uxnqqxsx8snjy":         "JUKI-0010",
	"11ufofajlpfv7el4x6wguwsv": "JUKI-0013",
	"21ctfe2mhhhqeqic":         "JUKI-0019",
	"214v7ufnqyeetbt9":         "JUITA-0034",
}

func normalizeMachineUUIDKey(uuid string) string {
	return strings.ToLower(strings.TrimSpace(uuid))
}
