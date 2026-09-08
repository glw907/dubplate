---
name: app-svelte-no-store-di
description: dubplate's web/src/App.svelte has no store-injection seam, unlike its child components
metadata:
  type: project
---

`web/src/App.svelte` imports the module-singleton `uploadStore` directly
(`import { uploadStore } from './lib/store/upload.svelte'`) and calls
`uploadStore.X` throughout. Unlike `History.svelte`/`Journey.svelte`,
which both accept a `store: UploadStore = uploadStore` prop for test
injection, App.svelte has no such prop.

**Why:** the plan/pass history never added one; App.svelte predates the
convention or was never revisited once History/Journey established it.

**How to apply:** an `App.test.ts` (added U13 fold F5) drives the real
singleton directly: stub `fetch` per test via `vi.stubGlobal`, and reset
every `uploadStore` field (`config`, `signedOut`, `files`, `journey`,
`journeyOutcome`, `history`, plus `resetSession()`) in `beforeEach`.
Vitest gives each test *file* its own module registry, so this reset is
safe and does not leak into other test files. If a future pass finds
this pattern too cramped (e.g. needs to test two independent App
instances in one file), the clean fix is adding a `store` prop to
App.svelte matching History/Journey's own convention, not a workaround
in the test file.

**Driving a mid-upload file state (F5 review round):** to test a phase
that only exists once real files are in flight (e.g. a server rejection
arriving on a file already uploading), do not let `startFile` run for
real -- it opens an actual `tus-js-client` upload via `TusDriver`
against the fetch-stubbed endpoint, which is untested territory and
risks flakiness/dangling handles. Instead `vi.spyOn(uploadStore,
'startFile').mockImplementation(() => {})` before rendering, drive the
UI far enough that `uploadStore.files` has the rows (addFile runs
synchronously inside `beginUpload`, so no `waitFor` is strictly needed,
though asserting the array length first is cheap insurance), then
mutate a row's `status`/`failureKind`/`verdict` directly -- `files` is a
class-level `$state` array, so this reactive-property mutation from
outside the component is exactly what the store's own internal methods
do (`find(id)` returns the same live entry) and it propagates correctly.
Restore the spy in that describe block's own `afterEach(() =>
vi.restoreAllMocks())`; the top-level `afterEach` in this file does not
call it, and a stray no-op `startFile` would silently break any
later-running test in the same file.
