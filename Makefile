SIM_SRC = ./sim/*.ail

# AILANG compilation (v0.5.0+)
# Compile all .ail files together for multi-module support
sim:
	ailang compile --emit-go --package-name sim_gen --out . sim/protocol.ail sim/world.ail sim/npc_ai.ail sim/step.ail

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

# Screenshot capture for AI self-testing
screenshot:
	@mkdir -p out
	go run ./cmd/game -screenshot 60 -output out/screenshot.png -seed 1234

screenshot-zoomed:
	@mkdir -p out
	go run ./cmd/game -screenshot 60 -output out/screenshot-zoomed.png -seed 1234 -camera 0,0,2.0

screenshot-panned:
	@mkdir -p out
	go run ./cmd/game -screenshot 60 -output out/screenshot-panned.png -seed 1234 -camera 200,200,1.0

screenshots: screenshot screenshot-zoomed screenshot-panned
	@echo "Screenshots saved to out/"

# Test scenarios for AI self-testing
scenario-pan:
	go run ./cmd/game -scenario camera-pan

scenario-zoom:
	go run ./cmd/game -scenario camera-zoom

scenario-npc:
	go run ./cmd/game -scenario npc-movement

scenarios: scenario-pan scenario-zoom scenario-npc
	@echo "Scenarios complete. Output in out/scenarios/"

# Visual regression testing (golden files)
test-visual:
	.claude/skills/test-manager/scripts/run_tests.sh

test-golden:
	.claude/skills/test-manager/scripts/compare_golden.sh

update-golden:
	.claude/skills/test-manager/scripts/update_golden.sh

# Testing and linting
test:
	go test -v ./...

test-all: test test-visual test-golden

lint:
	go vet ./...

# Clean (preserves sim_gen for mock mode)
clean:
	rm -rf bin out/*

clean-all:
	rm -rf sim_gen bin out/*

.PHONY: sim game eval run game-mock eval-mock run-mock sprites test test-all test-visual test-golden update-golden lint clean clean-all screenshot screenshot-zoomed screenshot-panned screenshots scenario-pan scenario-zoom scenario-npc scenarios
