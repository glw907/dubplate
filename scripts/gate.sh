#!/usr/bin/env bash
# scripts/gate.sh: run scripts/check.sh through cairn-run-gate.
# A worktree-isolated session's Bash guard refuses `cairn-run-gate '<string>'`,
# since it cannot show the quoted string is not git; a plain call to a script
# that execs cairn-run-gate passes, and keeps the lane lock and the run records
# `--records` scores. The wrapper accepts a fixed set of flags and builds the
# gate string from them alone, because cairn-run-gate hands its string to
# `bash -c`, and a forwarded free-form argument would reach that shell.
# Exit codes are cairn-run-gate's, so exit 75 still means re-issue.
#
# usage: scripts/gate.sh [--scope-changed] [--merge] [--dry-run] [--base <40-hex sha>]
set -uo pipefail
cd "$(dirname "$0")/.." || exit 2

usage='usage: scripts/gate.sh [--scope-changed] [--merge] [--dry-run] [--base <40-hex sha>]'

args=''
dry_run=0
while [ $# -gt 0 ]; do
	case "$1" in
	--scope-changed | --merge)
		args="${args:+$args }$1"
		shift
		;;
	--dry-run)
		dry_run=1
		args="${args:+$args }$1"
		shift
		;;
	--base)
		# An explicit digit list, never a range: a range matches by locale.
		if [ $# -lt 2 ] || [ "${#2}" -ne 40 ] || [[ "$2" == *[!0123456789abcdef]* ]]; then
			echo "$usage" >&2
			exit 2
		fi
		args="${args:+$args }--base $2"
		shift 2
		;;
	*)
		echo "$usage" >&2
		exit 2
		;;
	esac
done

# The full gate's rendered leg launches Chromium, so it takes the heavy lane
# (cairn-run-gate's default). A dry run launches no browser, so it queues on
# the light lane's lock instead of behind a heavy gate. Any other run clears an
# inherited lane, so the full gate never runs under the light lane's memory cap.
if [ "$dry_run" -eq 1 ]; then
	export CAIRN_GATE_LANE=light
else
	unset CAIRN_GATE_LANE
fi
exec cairn-run-gate "bash scripts/check.sh${args:+ $args}"
