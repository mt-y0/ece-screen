# ECE-screen — `make` shows the menu. Playful output lives in scripts/tui.sh.
BIN     := ECE-screen
STAMP   := .build/deps.stamp
TUI     := ./scripts/tui.sh
SRC     := $(shell find . -name '*.go' -not -path './pb_data/*') go.mod go.sum
export GOFLAGS ?= -buildvcs=false

.DEFAULT_GOAL := help
.PHONY: help setup deps doctor build run labwc test vet fmt clean distclean

help: ## show this menu
	@$(TUI) banner
	@awk 'BEGIN{FS=":.*## "} /^[a-z]+:.*## /{printf "  \033[38;5;213m%-10s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)
	@echo

doctor: ## check system dependencies (go, libmpv, ffprobe, ...)
	@$(TUI) say "checking the neighbourhood..."
	@$(TUI) doctor

deps: ## install missing system packages (Arch: go gcc mpv ffmpeg X11/GL libs labwc xwayland)
	@$(TUI) deps

# Fresh clone? (no stamp, or go.mod/go.sum/tui.sh changed) -> install system packages
# (Arch/pacman, via sudo), verify them, then fetch Go modules (raylib-go, go-mpv, PocketBase).
$(STAMP): go.mod go.sum scripts/tui.sh
	@$(TUI) banner
	@$(TUI) say "fresh download detected — checking dependencies"
	@$(TUI) deps
	@$(TUI) doctor
	@$(TUI) step "📦" "downloading Go modules" -- go mod download
	@mkdir -p .build && touch $@

setup: $(STAMP) ## download dependencies (auto-runs on first build/run)

$(BIN): $(STAMP) $(SRC)
	@$(TUI) step "🔨" "compiling $(BIN)" -- go build -o $(BIN) .

build: $(BIN) ## compile the binary (fetches deps on a fresh clone)
	@$(TUI) done "built ./$(BIN) — run it with: make run"

run: $(BIN) ## build if needed, then launch the board (needs display + libmpv)
	@$(TUI) say "launching the board — press X to quit 🚀"
	@./$(BIN)

vet: $(STAMP) ## go vet
	@$(TUI) step "🔍" "go vet" -- go vet ./...

labwc: $(BIN) ## build if needed, then run the board inside labwc (labwc -S)
	@command -v labwc >/dev/null 2>&1 || { $(TUI) fail "labwc not installed — run: make deps"; exit 1; }
	@$(TUI) say "starting labwc with the board (quit with X; labwc exits with it) 🚀"
	@exec labwc -S "$(CURDIR)/$(BIN)"

test: $(STAMP) ## vet + gofmt check + go test
	@$(TUI) step "🔍" "go vet" -- go vet ./...
	@$(TUI) step "🎨" "gofmt check" -- sh -c 'out=$$(gofmt -l $$(find . -name "*.go" -not -path "./pb_data/*")); [ -z "$$out" ] || { echo "unformatted:"; echo "$$out"; exit 1; }'
	@$(TUI) step "🧪" "go test" -- go test ./...
	@$(TUI) done "all checks passed"

fmt: ## gofmt -w all sources
	@$(TUI) step "🎨" "formatting" -- gofmt -w $$(find . -name '*.go' -not -path './pb_data/*')

clean: ## remove the binary and build stamp (keeps pb_data)
	@rm -rf $(BIN) .build
	@$(TUI) ok "squeaky clean 🧽"

distclean: clean ## clean + wipe Go build cache for this module
	@$(TUI) step "🧹" "cleaning go cache" -- go clean -cache
