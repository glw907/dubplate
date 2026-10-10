#!/usr/bin/env node
// The pass-execute shim: turns the runner's recorded range into the gate string to run.
// The runner fixes this path and calls it as
// `node scripts/checks/gate-tier.mjs --range <base>..HEAD [--paint yes|no] [--pin <tier>]`.

import { parseArgs } from "node:util";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..", "..");
const fullForm = `bash ${join(root, "scripts", "check.sh")}`;
const sha40 = /^[0-9a-f]{40}$/;

/**
 * Writes the decision to stderr, leaving stdout for the gate string alone.
 *
 * @param msg - One line of explanation.
 */
function note(msg) {
  process.stderr.write(`gate-tier: ${msg}\n`);
}

/**
 * Sizes the gate for a range. The base is shape-checked only so it is safe to embed in a shell
 * command; `check.sh` validates everything else and fails closed to the full form itself.
 *
 * @param range - The `--range` value, `<base>..<right>`.
 * @param pinned - Whether a `--pin` was given; every pin resolves to the full form.
 * @returns The gate string and the reason for it.
 */
function decide(range, pinned) {
  if (pinned) {
    return { gate: fullForm, why: "pinned" };
  }
  const cut = range.indexOf("..");
  const base = range.slice(0, cut);
  const right = range.slice(cut + 2);
  if (base === "") {
    return { gate: fullForm, why: "empty base" };
  }
  if (!sha40.test(base)) {
    return { gate: fullForm, why: "base is not a 40-hex commit" };
  }
  if (right !== "HEAD") {
    return { gate: fullForm, why: "range does not end at HEAD" };
  }
  return { gate: `${fullForm} --scope-changed --base ${base}`, why: "scoped to the range" };
}

const { values } = parseArgs({
  options: {
    range: { type: "string" },
    paint: { type: "string" },
    pin: { type: "string" },
    class: { type: "string" },
    protected: { type: "boolean" },
  },
  strict: false,
});

if (typeof values.range !== "string" || !values.range.includes("..")) {
  note("--range <base>..HEAD is required");
  process.exit(2);
}
if (values.paint !== undefined) {
  note("--paint ignored: dubplate has no paint tier, and any web/ change runs the rendered leg");
}
if (values.class !== undefined) {
  note("--class ignored: check.sh classifies by path");
}
if (values.protected) {
  note("--protected ignored: dubplate has no CI wait");
}
if (typeof values.pin === "string" && values.pin !== "full") {
  note(`--pin ${values.pin} has no narrower tier here; using the full form`);
}

const { gate, why } = decide(values.range, values.pin !== undefined);
note(why);
process.stdout.write(`${gate}\n`);
