# AILANG compilation (v0.5.5+)
# Just pass the directory - AILANG discovers .ail files automatically
sim:
	ailang compile --emit-go --package-name sim_gen --out . sim/

# Build targets (depend on sim when ailc is available)
game: sim
	go build -o bin/game ./cmd/game

eval: sim
	@mkdir -p out/eval
	go test -bench=. -benchmem ./engine/bench > out/eval/bench.txt
	go run ./cmd/eval

run: sim
	go run ./cmd/game

# Demo targets
demo-bridge: sim
	go run ./cmd/demo-bridge --debug

# CLI tool
cli:
	go build -o bin/voyage ./cmd/voyage

cli-mock:
	go build -o bin/voyage ./cmd/voyage

# Install voyage CLI globally
install:
	go install ./cmd/voyage

# Performance testing with threshold checks
perf: cli
	@mkdir -p out/eval
	./bin/voyage perf -o out/eval/perf.json

perf-ci: cli
	@mkdir -p out/eval
	./bin/voyage perf -o out/eval/perf.json -fail=true

# Mock targets (use existing sim_gen, no AILANG compiler needed)
game-mock:
	go build -o bin/game ./cmd/game

eval-mock:
	@mkdir -p out/eval
	go test -bench=. -benchmem ./engine/bench > out/eval/bench.txt 2>&1 || true
	go run ./cmd/eval

run-mock:
	go run ./cmd/game

# Generate test sprites
sprites:
	go run ./cmd/gensprites

# Screenshot capture for AI self-testing
screenshot:
	@mkdir -p out/screenshots
	go run ./cmd/game -screenshot 60 -output out/screenshots/screenshot.png -seed 1234

screenshot-zoomed:
	@mkdir -p out/screenshots
	go run ./cmd/game -screenshot 60 -output out/screenshots/screenshot-zoomed.png -seed 1234 -camera 0,0,2.0

screenshot-panned:
	@mkdir -p out/screenshots
	go run ./cmd/game -screenshot 60 -output out/screenshots/screenshot-panned.png -seed 1234 -camera 200,200,1.0

screenshots: screenshot screenshot-zoomed screenshot-panned
	@echo "Screenshots saved to out/screenshots/"

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

# =============================================================================
# BUILD TARGETS WITH sim_gen ERROR DETECTION
# =============================================================================
# IMPORTANT: Claude should use these instead of direct `go build`
# If errors occur in sim_gen/, it's an AILANG codegen bug that must be reported

# Build all Go code with sim_gen error detection
build: sim
	@echo "Building with sim_gen error detection..."
	@go build ./... > /tmp/go-build.log 2>&1; \
	BUILD_STATUS=$$?; \
	cat /tmp/go-build.log; \
	if [ $$BUILD_STATUS -eq 0 ]; then \
		echo "✓ Build successful"; \
	else \
		if grep -q 'sim_gen/' /tmp/go-build.log; then \
			echo ""; \
			echo "═══════════════════════════════════════════════════════════════════════"; \
			echo "❌ ERROR IN sim_gen/ - THIS IS AN AILANG CODEGEN BUG"; \
			echo "═══════════════════════════════════════════════════════════════════════"; \
			echo ""; \
			echo "DO NOT try to work around this error!"; \
			echo ""; \
			echo "1. Report the bug:"; \
			echo '   ailang messages send user "AILANG codegen bug: <paste error>" \'; \
			echo '     --title "Codegen: <brief description>" \'; \
			echo '     --from "stapledons_voyage" \'; \
			echo '     --type bug \'; \
			echo '     --github'; \
			echo ""; \
			echo "2. Mark the feature as BLOCKED in sprint tracking"; \
			echo ""; \
			echo "3. WAIT for AILANG team to fix it"; \
			echo ""; \
			echo "═══════════════════════════════════════════════════════════════════════"; \
			exit 1; \
		else \
			echo "Build error (not in sim_gen - normal Go error to fix)"; \
			exit 1; \
		fi; \
	fi

# Build only engine code (excludes sim_gen from the check - for engine-only work)
engine:
	@echo "Building engine (sim_gen excluded from error check)..."
	go build ./engine/... ./cmd/...

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

.PHONY: sim game eval run demo-bridge cli cli-mock install game-mock eval-mock run-mock sprites build engine test test-all test-visual test-golden update-golden lint clean clean-all screenshot screenshot-zoomed screenshot-panned screenshots scenario-pan scenario-zoom scenario-npc scenarios
