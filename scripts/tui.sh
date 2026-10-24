#!/usr/bin/env bash
# Playful terminal UI helpers for the Makefile. Usage: scripts/tui.sh <cmd> [args]
#   banner | step <emoji> <title> -- <cmd...> | say <msg> | ok <msg> | warn <msg> | fail <msg>
#   doctor | deps | done <msg>
if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  R=$'\e[0m'; B=$'\e[1m'; D=$'\e[2m'
  C1=$'\e[38;5;213m'; C2=$'\e[38;5;117m'; C3=$'\e[38;5;227m'; C4=$'\e[38;5;120m'; C5=$'\e[38;5;203m'
  TTY=1
else
  R=; B=; D=; C1=; C2=; C3=; C4=; C5=; TTY=0
fi

rainbow() { # print text with cycling colours
  local s="$1" i c out= cols=(213 177 117 123 120 227 215 203)
  for ((i = 0; i < ${#s}; i++)); do
    c=${s:i:1}
    if [ "$TTY" = 1 ]; then out+=$'\e[38;5;'"${cols[i % ${#cols[@]}]}"'m'"$c"; else out+=$c; fi
  done
  printf '%s%s\n' "$out" "$R"
}

banner() {
  echo
  rainbow "   ╭─────────╮"
  rainbow "   │  ◕   ◕  │   E C E · S C R E E N"
  rainbow "   │    ‿    │   the department display board"
  rainbow "   ╰─┬─────┬─╯   60fps · raylib · pocketbase"
  echo
}

say()  { printf '  %s✦%s %s\n' "$C2" "$R" "$*"; }
ok()   { printf '  %s✔%s %s\n' "$C4" "$R" "$*"; }
warn() { printf '  %s⚠%s %s\n' "$C3" "$R" "$*"; }
fail() { printf '  %s✘%s %s\n' "$C5" "$R" "$*"; }

# step <emoji> <title> -- cmd...  : runs cmd with a spinner, shows log on failure
step() {
  local emoji=$1 title=$2; shift 3
  local log; log=$(mktemp)
  if [ "$TTY" = 0 ]; then
    echo ">> $title"
    "$@" >"$log" 2>&1; local rc=$?
    [ $rc -ne 0 ] && cat "$log"
    rm -f "$log"; return $rc
  fi
  local frames=(⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏) i=0 start=$SECONDS
  "$@" >"$log" 2>&1 &
  local pid=$!
  printf '\e[?25l'
  while kill -0 "$pid" 2>/dev/null; do
    printf '\r  %s%s%s %s %s%s%s %s(%ds)%s\e[K' "$C1" "${frames[i++ % 10]}" "$R" "$emoji" "$B" "$title" "$R" "$D" $((SECONDS - start)) "$R"
    sleep 0.08
  done
  wait "$pid"; local rc=$?
  printf '\e[?25h\r\e[K'
  if [ $rc -eq 0 ]; then
    printf '  %s✔%s %s %s %s(%ds)%s\n' "$C4" "$R" "$emoji" "$title" "$D" $((SECONDS - start)) "$R"
  else
    printf '  %s✘%s %s %s\n' "$C5" "$R" "$emoji" "$title"
    printf '%s' "$D"; tail -n 40 "$log" | sed 's/^/    │ /'; printf '%s' "$R"
  fi
  rm -f "$log"; return $rc
}

doctor() { # check system dependencies; exit 1 if a required one is missing
  local bad=0
  chk() { # name required hint cmd...
    local name=$1 req=$2 hint=$3; shift 3
    if "$@" >/dev/null 2>&1; then ok "$name"
    elif [ "$req" = req ]; then fail "$name ${D}→ $hint${R}"; bad=1
    else warn "$name ${D}→ $hint${R}"; fi
  }
  chk "go toolchain"            req "install Go (https://go.dev/dl)"                     command -v go
  chk "libmpv (pkg-config)"     req "Arch: pacman -S mpv · Debian: apt install libmpv-dev" pkg-config --exists mpv
  chk "C compiler (cgo)"       req "Arch: pacman -S gcc · Debian: apt install build-essential" command -v gcc
  chk "X11/GL dev libs"         req "Arch: pacman -S libx11 libxrandr libxinerama libxcursor libxi libglvnd · Debian: apt install libx11-dev libxrandr-dev libxinerama-dev libxcursor-dev libxi-dev libgl-dev" sh -c 'pkg-config --exists x11 xrandr xinerama xcursor xi gl'
  chk "ffprobe (video sizing)"  opt "Arch: pacman -S ffmpeg · Debian: apt install ffmpeg" command -v ffprobe
  chk "labwc (make labwc)"      opt "Arch: pacman -S labwc xorg-xwayland"                 command -v labwc
  chk "twemoji PNGs"            opt "resources/fonts/twemoji missing — emoji won't render" test -d resources/fonts/twemoji
  chk "display (\$DISPLAY/\$WAYLAND_DISPLAY)" opt "needed only for 'make run'"            test -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}"
  return $bad
}

# deps: make sure the system packages needed to build and run exist.
# Arch (pacman): installs whatever is missing with sudo. Other distros: lists what's missing.
# DEPS_DRY_RUN=1 prints the install command instead of running it.
deps() {
  local missing=()
  need() { # pkg check...   (adds pkg when the check fails)
    local pkg=$1; shift
    "$@" >/dev/null 2>&1 || missing+=("$pkg")
  }
  pc() { command -v pkg-config >/dev/null 2>&1 && pkg-config --exists "$@"; }

  need go          command -v go
  need gcc         command -v gcc                # raylib-go compiles raylib via cgo
  need pkgconf     command -v pkg-config
  need mpv         pc mpv                        # libmpv
  need ffmpeg      command -v ffprobe            # video sizing
  need libx11      pc x11                        # raylib/GLFW build + runtime
  need libxrandr   pc xrandr
  need libxinerama pc xinerama
  need libxcursor  pc xcursor
  need libxi       pc xi
  need libglvnd    pc gl
  need labwc       command -v labwc              # `make labwc` compositor
  need xorg-xwayland command -v Xwayland         # lets the X11 window run inside labwc

  if [ ${#missing[@]} -eq 0 ]; then ok "system packages present"; return 0; fi

  if ! command -v pacman >/dev/null 2>&1; then
    warn "missing packages (install with your package manager): ${missing[*]}"
    return 0 # `doctor` decides whether anything required is really absent
  fi

  say "installing missing packages: ${missing[*]}"
  local sudo=sudo
  [ "$(id -u)" = 0 ] && sudo=
  if [ -n "$sudo" ] && ! command -v sudo >/dev/null 2>&1; then
    fail "sudo not found — run as root: pacman -S --needed ${missing[*]}"; return 1
  fi
  if [ "${DEPS_DRY_RUN:-}" = 1 ]; then
    echo "DRY RUN: $sudo pacman -S --needed --noconfirm ${missing[*]}"; return 0
  fi
  # Runs on the real terminal (no spinner) so sudo can prompt for a password.
  if $sudo pacman -S --needed --noconfirm "${missing[@]}"; then ok "packages installed"
  else fail "pacman failed — install manually: ${missing[*]}"; return 1; fi
}

finish() { echo; rainbow "  ★ $* ★"; echo; }

cmd=${1:-banner}; shift || true
case $cmd in
  banner) banner ;; say) say "$@" ;; ok) ok "$@" ;; warn) warn "$@" ;; fail) fail "$@" ;;
  step) step "$@" ;; doctor) doctor ;; deps) deps ;; done) finish "$@" ;;
  *) echo "unknown: $cmd" >&2; exit 2 ;;
esac
