package main

import (
	"encoding/json"
	"os"

	"stapledons_voyage/engine/scenario"
)

func main() {
	report := scenario.RunAll()

	f, err := os.Create("out/report.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		panic(err)
	}
}
