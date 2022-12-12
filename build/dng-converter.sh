#!/usr/bin/env bash
set -euo pipefail

ProgramW6432=$(wine cmd /c echo '%ProgramW6432%' | tr -d '\r')

args=()
for arg in "$@"; do
  if [ -f "$arg" ]; then
    args+=($(winepath -w "$arg"))
  else
    args+=("$arg")
  fi
done

exec wine "$ProgramW6432"'\Adobe\Adobe DNG Converter\Adobe DNG Converter.exe' "${args[@]}"