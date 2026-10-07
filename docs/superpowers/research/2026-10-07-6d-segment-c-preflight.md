# 6d segment C pre-flight (2026-10-07, HEAD after Tasks 1-5)

Every behavioral claim and quoted behavior in Tasks 6-8 holds. Only line numbers moved, plus one signature.

## Changed signature (Task 6)
- `newServeShell` (serve.go:303) is now `(ctx, environ, reg, prefix, bindAddrs, cfg, authCfg, logger) (*serve.Shell, *store.Handle, error)`. Inside it `openServeStore` (:308) and `applySettingsSeed` (:317) run before `auth.Open`. `newServeCmd` (:238) still calls `resolveAuthConfig(environ)` before `newServeShell`.
- So Task 6's "read all ten values once after the seed through the accessor over its handle" happens inside `newServeShell`, after `applySettingsSeed`, where `h` exists. `authCfg` can no longer be a plain parameter computed before the store opens, or the test callers change shape; name the choice in the report.

## Moved pointers
| Plan cites | Now at |
|---|---|
| auth.go:518-563 resolveAuthConfig | auth.go:527-572 |
| auth.go authConfig :499-513 | ~:498-519 |
| auth.go validateMagicLinkSource :265-338 | :274-~350 |
| auth.go relational checks :303-322 | :314-331 (retention below expiry :321, lifetime above cap :329-330) |
| auth_test.go:830-890 | TestResolveAuthConfig :832-~890 |
| serve_test.go resolveAuthConfig callers :77, :603, :696 | :85, :585, :721 |
| serve_test.go newServeShell caller :596-610 | ~:585-600 |
| rendered_test.go checkRateBudget :494-500 | :480-484 |
| rendered_test.go mintLinkFor :541-545 | :526-529 |
| store.go settingShown :333-338 | :421-423 |
| store.go "serve's next start" :290/:292/:294/:307 | :378, :380, :382, :395 |
| store.go settingsDoctorRows :226-252 | :314 |
| store.go settingsNoTableStatus :302-310 | :392 |
| store.go storeRow / storeVouches | :222 / :279 |
| check.go evaluation behind r.Facts != nil :62-68 | still guarded; printSettingsRows ~:86 |
| backup.go logSettingsEvaluation | ~:262-277, "next start" Warn :274; settings.Evaluate :222 |
| settings/evaluate.go EvaluateFacts :128-148 | :134 |
| settings/evaluate.go evaluateTableAbsent | :150 (still "never reads environ") |
| evaluate.go parse | :368 (still no KindString arm) |
| Outcome.Report sensitive branch | evaluate.go:347 |
| auth/session.go VerifySession :213-219 | :219 |
| auth/seam.go Lookup :112 | :112 (identity now carries UserID: member.ID) |
| serve.go sweepNonces :485-489 | :541 |
| serve.go authHandlersConfig | :482 |
| serve.go OpenNonceStore :320 | :330 |
| serve.go NewLimiter :334-339 | :338 |
| open_test.go:633 | :634 |

## Holding (no change)
- Store-test hint sites leaf_test.go:129,:179; inspect_test.go:182; open_test.go:634; restore_test.go:128-134; leaf.go:113 runuser hint.
- unidecode.go:258,:269; unidecode_test.go:141-142,:198-199; verb reads DUBPLATE_STATE_DIR at :82,:147; comment :111.
- Exactly ten auth thresholds in registerAuthConfig. magicLinkSource.Required has eight keys (Task 5 removed the roster key; env template says "eight").
- AddCorrelatedRule registry.go:461 panics on an unregistered setting, so register the ten settings first. claim :125, AddSetting :431, AddSuperseded :502.
- config/settings.go:58 "KindString is refused"; Sensitive :69-70 "must not print".
- buildValidator panics :129; boundedValidator :139,:143,:146.
- internal/auth does not import settings. EpochStore.Bump (epoch.go:103) has no non-test caller (R-6d20 holds).
- Settings accessor has Int, Duration, Float only (accessor.go:43-53).
- nonce.go Insert :149, Sweep :200 (no ctx); RunSweep :123 takes sweep func(now time.Time); linkExpiryCeiling doc :69-76.
- handlers.go LinkExpiry/SessionLifetime/SessionAbsoluteCap :93-98, BodyCap :111, GenericAnswer :118, Lookup :209,:365.
- Task 8 files unchanged since 0cb33ae: systemd/dubplate-server-serve.service (ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/roster at :43); deploy.sh header leaves :29-30, DUBPLATE_SERVER_LEAVES=(auth roster) :281, loop :282, leaf-table comment :268-280, unidecode comment :287-298; check.sh non-reduced block :626-737, state-dir leg :657-737; runbook roster mentions :75-95, :136-137, :162, re-create section :216-250, doctor roster row :265, settings rows :332-345 ("not yet true" text), "roster beside it" :773.

## Already done by Tasks 1-5 (do not redo)
- Env template's DUBPLATE_ROSTER_FILE row gone; header says "eight required keys".
- AddSuperseded for DUBPLATE_ROSTER_FILE at auth.go:236.
- doctor's members row exists (doctor.go:227-255); the runbook still documents a `roster` row, which Task 8 rewrites.
