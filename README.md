Stapledons Voyage – Game Repository Bootstrap Brief

(for integrating with AILANG as the sim layer)

This repository hosts the game engine + assets + evaluation harness, while AILANG (in a separate repo) provides the simulation logic and AI behaviour.

The game repo is deliberately thin: it compiles AILANG → Go, links the generated code into an Ebiten-based engine, and runs benchmarks + scenarios.

Think of this repo as the “host”, and AILANG as the “brain”.

⸻

1. Repository Structure

stapledon/
│
├── cmd/
│   ├── game/                 # Ebiten-powered game runtime (main window loop)
│   │   └── main.go
│   └── eval/                 # Benchmark + scenario runner CLI
│       └── main.go
│
├── sim/                      # AILANG source code
│   ├── world.ail             # World state definitions
│   ├── protocol.ail          # FrameInput/FrameOutput/DrawCmd (stable API)
│   ├── step.ail              # init_world + step implementation
│   └── npc_ai.ail            # NPC AI logic using AI effect
│
├── sim_gen/                  # Auto-generated Go code from AILANG
│   └── (generated *.go files)
│
├── engine/                   # Pure Go code
│   ├── render/               # Ebiten draw bridge
│   │   ├── assets.go
│   │   ├── draw.go
│   │   └── input.go
│   ├── scenario/
│   │   ├── definitions/      # YAML/JSON scenarios
│   │   ├── runner.go
│   │   └── metrics.go
│   └── bench/
│       └── bench_test.go
│
├── assets/                   # Sprites, tilesheets, fonts
│
├── out/
│   ├── report.json           # Generated eval output
│   └── screenshots/          # Optional captured frames
│
├── Makefile
└── go.mod

This mirrors standard Go project layouts and keeps the AILANG-generated code strictly separated in sim_gen/.

⸻

2. Build Flow

Step 1: Compile AILANG → Go

You run:

make sim

Which internally runs:

ailc --emit-go --package-name sim_gen --out ./sim_gen ./sim/*.ail

After this step:
	•	You have a Go package planetworld/sim_gen.
	•	It contains:
	•	InitWorld
	•	Step
	•	All ADTs (World, FrameInput, FrameOutput, DrawCmd, …)
	•	You don’t modify these files manually.

Step 2: Build engine + game

make game

This runs:

go build -o bin/game ./cmd/game

Result: a native executable (./bin/game) that opens a window.

Step 3: Run evaluation

make eval

Which runs:
	•	Benchmarks (go test ./engine/bench -bench=. -benchmem)
	•	Scenario suite (go run ./cmd/eval)
	•	Combines them into out/report.json

This JSON is what you paste into ChatGPT to guide AILANG improvements.

⸻

3. Initial AILANG Code

Inside sim/ you start with three core files:

protocol.ail

Defines stable boundary types:

module game/protocol

type Coord = { x: int, y: int }

type MouseState = {
    x: float,
    y: float,
    buttons: List<int>
}

type KeyEvent = {
    key: int,
    kind: string    -- "down" | "up"
}

type FrameInput = {
    mouse: MouseState,
    keys: List<KeyEvent>,
}

type DrawCmd =
  | Sprite { id: int, x: float, y: float, z: int }
  | Rect   { x: float, y: float, w: float, h: float, color: int, z: int }
  | Text   { text: string, x: float, y: float, z: int }

type FrameOutput = {
    draw: List<DrawCmd>,
    sounds: List<int>,
    debug: List<string>
}

world.ail

Minimal world:

module game/world
import game/protocol (Coord)

type Tile = { biome: int }

type PlanetState = {
    width: int,
    height: int,
    tiles: Array<Tile>,
}

type NPC = {
    id: int,
    pos: Coord,
}

type World = {
    tick: int,
    planet: PlanetState,
    npcs: List<NPC>,
}

step.ail

Top-level sim logic:

module game/step
import game/world (World, PlanetState, NPC)
import game/protocol (FrameInput, FrameOutput, DrawCmd)
import std/random (rand_int)
import std/list (map)

export func init_world(seed: int) -> World ! {RNG} {
    let w = 64
    let h = 64
    let tiles = Array.generate(w * h, \i. { biome: rand_int(4) })
    {
        tick: 0,
        planet: { width: w, height: h, tiles },
        npcs: [],
    }
}

export func step(world: World, input: FrameInput)
  -> (World, FrameOutput) ! {RNG} {

    let newTick = world.tick + 1

    let drawTiles =
        Array.to_list(
            Array.generate(
                world.planet.width * world.planet.height,
                \i.
                    let x = i % world.planet.width
                    let y = i / world.planet.width
                    DrawCmd.Rect { x: float(x*4), y: float(y*4),
                                   w: 4.0, h: 4.0,
                                   color: world.planet.tiles[i].biome,
                                   z: 0 }
            )
        )

    let out = {
        draw: drawTiles,
        sounds: [],
        debug: [],
    }

    ({ world with tick: newTick }, out)
}

This is “dumb but real”:
It renders coloured tiles and increments ticks — enough to confirm the pipeline.

⸻

4. Go Engine Bootstrap

Inside cmd/game/main.go:

package main

import (
    "log"
    "github.com/hajimehoshi/ebiten/v2"
    "planetworld/sim_gen"
    "planetworld/engine/render"
)

type Game struct {
    world sim_gen.World
    out   sim_gen.FrameOutput
}

func (g *Game) Update() error {
    input := render.CaptureInput()
    w2, out, err := sim_gen.Step(g.world, input)
    if err != nil { return err }
    g.world = w2
    g.out = out
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    render.RenderFrame(screen, g.out)
}

func (g *Game) Layout(w,h int) (int,int) {
    return 640, 480
}

func main() {
    w := sim_gen.InitWorld(1234)
    game := &Game{world: w}

    ebiten.SetWindowTitle("PlanetWorld")
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}

And inside engine/render/ you implement:
	•	CaptureInput()
Convert Ebiten mouse/keyboard into FrameInput.
	•	RenderFrame(screen, FrameOutput)
Switch on DrawCmd variants and draw rectangles, sprites, and text.

⸻

5. Evaluation Harness Bootstrap

Inside cmd/eval/main.go:

package main

import (
    "encoding/json"
    "os"
    "planetworld/engine/scenario"
    "planetworld/engine/bench"
)

func main() {
    report := scenario.RunAll()
    bench.AttachBenchmarks(report)

    f, _ := os.Create("out/report.json")
    _ = json.NewEncoder(f).Encode(report)
}

And a simple report struct:

type Report struct {
    Benchmarks map[string]BenchResult     `json:"benchmarks"`
    Scenarios  []scenario.Result           `json:"scenarios"`
}

This gives you the JSON needed for AI iteration.

⸻

6. Makefile Bootstrap

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


⸻

7. How This Repo Interacts With AILANG

The game repo:
	•	Does not modify AILANG codegen or the compiler.
	•	Does act as a consumer:
	•	If AILANG breaks ADT → Go mapping, the game breaks.
	•	If AILANG slows down, benchmarks report regressions.
	•	If AILANG effects misbehave, NPCs fail scenarios.

The relationship:
	•	The game becomes the primary integration test for AILANG.
	•	AILANG developers use:
	•	planetworld/out/report.json
	•	Sim crashes
	•	Render anomalies
to drive AILANG runtime/compiler improvements.

This is the feedback loop that builds both systems.

⸻

8. Summary

This bootstrap gives you:
	•	A clean folder structure
	•	AILANG compilation flow
	•	Ebiten-based engine skeleton
	•	Benchmark + scenario harness hooks
	•	Minimal AILANG sim starting point

You can now:
	•	Clone AILANG as a submodule or separate repo.
	•	Start iterating in sim/*.ail.
	•	Use make run to see changes visually.
	•	Use make eval to produce JSON reports for AI-based improvement.

⸻

If you want, I can also prepare:
	•	A planetworld logo banner
	•	A first scenario (single_colony_growth.yaml)
	•	A Tile rendering mock (ASCII)
	•	A CI config (GitHub Actions)

Just say the word.