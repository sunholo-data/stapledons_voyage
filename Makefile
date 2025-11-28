SIM_SRC = ./sim/*.ail

sim:
	ailc --emit-go --package-name sim_gen --out ./sim_gen $(SIM_SRC)

game: sim
	go build -o bin/game ./cmd/game

eval: sim
	go test -bench=. -benchmem ./engine/bench > out/bench.txt
	go run ./cmd/eval > out/report.json

run: sim
	go run ./cmd/game

clean:
	rm -rf sim_gen bin out/*

.PHONY: sim game eval run clean
