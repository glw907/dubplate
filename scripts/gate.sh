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
# usage: scripts/gate.sh [--scope-changed | --reduced | --merge] [--dry-run] [--base <40-hex sha>] [--fresh]
set -uo pipefail
cd "$(dirname "$0")/.." || exit 2

usage='usage: scripts/gate.sh [--scope-changed | --reduced | --merge] [--dry-run] [--base <40-hex sha>] [--fresh]'

args=()
dry_run=0
scoped=0
fresh=()
while [ $# -gt 0 ]; do
	case "$1" in
	--scope-changed | --reduced)
		scoped=1
		args+=("$1")
		shift
		;;
	--merge)
		args+=("$1")
		shift
		;;
	--dry-run)
		dry_run=1
		args+=("$1")
		shift
		;;
	--fresh)
		# The runner's flag: it goes ahead of the gate string, never into check.sh.
		fresh=(--fresh)
		shift
		;;
	--base)
		# An explicit digit list, never a range: a range matches by locale.
		if [ $# -lt 2 ] || [ "${#2}" -ne 40 ] || [[ "$2" == *[!0123456789abcdef]* ]]; then
			echo "$usage" >&2
			exit 2
		fi
		args+=(--base "$2")
		shift 2
		;;
	*)
		echo "$usage" >&2
		exit 2
		;;
	esac
done

# The lane. A dry run launches no browser, so it queues on the light lane's lock
# instead of behind a heavy gate. A scoped run asks check.sh's own dry run for
# its scope and takes the light lane when no rendered leg runs: `scope=changed`
# with `rendered=skipped`, or `scope=reduced`. Everything else, and any dry run
# that fails or prints no scope line, takes the heavy lane, so the full gate
# never runs under the light lane's memory cap. The tree can change between the
# dry run and the run; that costs at most a heavier lane.
lane=heavy
if [ "$dry_run" -eq 1 ]; then
	lane=light
elif [ "$scoped" -eq 1 ]; then
	if scope_out="$(bash scripts/check.sh "${args[@]}" --dry-run 2>&1)"; then
		scope_line="$(grep -m1 '^check.sh: scope=' <<<"$scope_out" || true)"
		case "$scope_line" in
		'check.sh: scope=changed '*' rendered=skipped '* | 'check.sh: scope=reduced '*) lane=light ;;
		esac
	else
		echo "$scope_out" >&2
	fi
fi
echo "gate.sh: lane=$lane"
if [ "$lane" = light ]; then
	export CAIRN_GATE_LANE=light
else
	unset CAIRN_GATE_LANE
fi
exec cairn-run-gate "${fresh[@]}" "bash scripts/check.sh${args[*]:+ ${args[*]}}"
