SIM_SRC = ./sim/*.ail

# AILANG compilation (requires ailc)
sim:
	ailc --emit-go --package-name sim_gen --out ./sim_gen $(SIM_SRC)

# Build targets (depend on sim when ailc is available)
game: sim
	go build -o bin/game ./cmd/game

eval: sim
	@mkdir -p out
	go test -bench=. -benchmem ./engine/bench > out/bench.txt
	go run ./cmd/eval > out/report.json

run: sim
	go run ./cmd/game

# Mock targets (use existing sim_gen, no AILANG compiler needed)
game-mock:
	go build -o bin/game ./cmd/game

eval-mock:
	@mkdir -p out
	go test -bench=. -benchmem ./engine/bench > out/bench.txt 2>&1 || true
	go run ./cmd/eval > out/report.json

run-mock:
	go run ./cmd/game

# Generate test sprites
sprites:
	go run ./cmd/gensprites

# Testing and linting
test:
	go test -v ./...

lint:
	go vet ./...

# Clean (preserves sim_gen for mock mode)
clean:
	rm -rf bin out/*

clean-all:
	rm -rf sim_gen bin out/*

.PHONY: sim game eval run game-mock eval-mock run-mock sprites test lint clean clean-all
