package main

import (
	"encoding/json"
	"os"

	"stapledons_voyage/engine/scenario"
)

func main() {
	report := scenario.RunAll()

	// Ensure output directory exists
	if err := os.MkdirAll("out/eval", 0755); err != nil {
		panic(err)
	}

	f, err := os.Create("out/eval/report.json")
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
