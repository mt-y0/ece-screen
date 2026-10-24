#!/usr/bin/env bash
# Playful terminal UI helpers for the Makefile. Usage: scripts/tui.sh <cmd> [args]
#   banner | step <emoji> <title> -- <cmd...> | say <msg> | ok <msg> | warn <msg> | fail <msg>
#   doctor | done <msg>
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
  chk "ffprobe (video sizing)"  opt "Arch: pacman -S ffmpeg · Debian: apt install ffmpeg" command -v ffprobe
  chk "twemoji PNGs"            opt "resources/fonts/twemoji missing — emoji won't render" test -d resources/fonts/twemoji
  chk "display (\$DISPLAY/\$WAYLAND_DISPLAY)" opt "needed only for 'make run'"            test -n "${DISPLAY:-}${WAYLAND_DISPLAY:-}"
  return $bad
}

finish() { echo; rainbow "  ★ $* ★"; echo; }

cmd=${1:-banner}; shift || true
case $cmd in
  banner) banner ;; say) say "$@" ;; ok) ok "$@" ;; warn) warn "$@" ;; fail) fail "$@" ;;
  step) step "$@" ;; doctor) doctor ;; done) finish "$@" ;;
  *) echo "unknown: $cmd" >&2; exit 2 ;;
esac
