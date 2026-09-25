/**
 * Proves the rendered test ran. `go test` exits 0 with "no tests to run"
 * when `-run` selects nothing, so a misspelled build tag or a renamed
 * test would pass the leg silently. This reads the `go test -json`
 * stream at argv[2], replays every event's output so the transcript
 * reads as `go test -v` would, and exits 1 unless one event passed the
 * top-level test itself: `Action` "pass" with `Test` exactly
 * `TestRendered`, never a subtest.
 */
import { readFileSync } from 'node:fs';
import { fail, isMainModule } from '../gates/lib.ts';

const TEST_NAME = 'TestRendered';

/** The `go doc cmd/test2json` TestEvent fields this reader uses. */
interface TestEvent {
  Action?: string;
  Test?: string;
  Output?: string;
}

function main(): void {
  const path = process.argv[2];
  if (!path) fail('test-events: usage: node scripts/rendered/test-events.ts <go test -json output>');
  let matched: string | null = null;
  for (const line of readFileSync(path, 'utf8').split('\n')) {
    if (line.trim() === '') continue;
    let event: TestEvent;
    try {
      event = JSON.parse(line) as TestEvent;
    } catch {
      process.stdout.write(`${line}\n`);
      continue;
    }
    if (event.Output !== undefined) process.stdout.write(event.Output);
    if (event.Action === 'pass' && event.Test === TEST_NAME) matched = line;
  }
  if (matched === null) {
    fail(`test-events: no "pass" event for ${TEST_NAME} in the go test -json stream; the rendered test did not run or did not pass.`);
  }
  process.stdout.write(`test-events: matched ${matched}\n`);
}

if (isMainModule(import.meta.url)) {
  main();
}
