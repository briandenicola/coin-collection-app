# Tasks: Table / Lightbox Showcase Mode

## Phase 1: Browser API lifecycle

- [x] T001 Add `useReducedMotion()` for `prefers-reduced-motion` with listener cleanup.
- [x] T002 Add `useFullscreen()` for target fullscreen entry/exit, state sync, and cleanup.
- [x] T003 Add `useWakeLock()` for progressive screen wake lock, release, visibility reacquire, and cleanup.
- [x] T004 Add composable regression tests for unsupported/supported lifecycle paths.

## Phase 2: Presentation viewer

- [x] T005 Add reusable `PresentCoinViewer.vue` for one-coin immersive presentation.
- [x] T006 Render dark full-bleed edge-lit stage with a single maximized image.
- [x] T007 Add tap-to-toggle metadata overlay using only name, ruler, denomination, era, material, and grade.
- [x] T008 Add obverse/reverse controls without pricing, value, notes, AI analysis, or owner fields.
- [x] T009 Add swipe and keyboard navigation events plus accessible exit/prev/next controls.
- [x] T010 Respect reduced motion for image and overlay transitions.

## Phase 3: Route and gallery launch

- [x] T011 Add authenticated `/present` route.
- [x] T012 Add `PresentModePage.vue` using current `store.coins` and `store.galleryIndex` as the source set.
- [x] T013 Add desktop collection header Present button.
- [x] T014 Add PWA collection menu Present action.
- [x] T015 Ensure exit returns to the collection route with existing filter/scroll state preserved by current route/store behavior.

## Phase 4: Tests and validation

- [x] T016 Test metadata allowlist and overlay toggle.
- [x] T017 Test swipe/keyboard navigation and obverse/reverse behavior.
- [x] T018 Test fullscreen/wake-lock cleanup on page exit/unmount.
- [x] T019 Test collection header launch wiring.
- [x] T020 Run targeted tests: `npm.cmd test -- PresentCoinViewer PresentMode useWakeLock useFullscreen CollectionPage PwaCollectionHeader --run`.
- [x] T021 Run full frontend tests: `npm.cmd test`.
- [x] T022 Run `npm.cmd run type-check`.
- [x] T023 Run `npm.cmd run build`.
- [x] T024 Run `npm.cmd run lint`.
- [x] T025 Confirm no backend/API changes or generated `dist/` artifacts are in the final diff.
