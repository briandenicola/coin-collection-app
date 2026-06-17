# Tasks: 3D Flip / Gyroscope Coin Viewer

**Input**: `specs/222-3d-coin-viewer/spec.md`, `specs/222-3d-coin-viewer/plan.md`  
**Branch**: `222-3d-coin-viewer-impl`  
**Target merge**: `beta`

## Phase 1: Foundation Helpers

- [x] T001 [P] Add `src/web/src/composables/useReducedMotion.ts` with a typed `prefersReducedMotion` readonly ref and `matchMedia('(prefers-reduced-motion: reduce)')` cleanup.
- [x] T002 [P] Add `src/web/src/composables/useDeviceOrientation.ts` with typed support detection, iOS permission request, RAF-throttled tilt state, clamping to +/-15 degrees, and unmount cleanup.
- [x] T003 [P] Add `src/web/src/composables/__tests__/useDeviceOrientation.test.ts` covering unsupported browser, permission granted, permission denied, clamped orientation updates, and listener cleanup.

## Phase 2: Reusable Coin Viewer

- [x] T004 Add `src/web/src/components/coin/CoinViewer3D.vue` with typed props for obverse/reverse source, alt text, size, interactivity, and tilt.
- [x] T005 Implement circular front/back faces with CSS 3D `rotateY`, rim/bevel treatment, token-based chrome, and no rectangular image reveal.
- [x] T006 Add an accessible flip button using a lucide icon and text/aria-label; disable it for one-image and no-image coins.
- [x] T007 Implement reduced-motion fallback so face changes do not animate and gyro tilt is disabled.
- [x] T008 Implement progressive orientation tilt/glint using `useDeviceOrientation()` only when `enableTilt` and reduced motion is false.
- [x] T009 Emit `flip` after a face change and `open-image` with the current face when the viewer body is clicked.
- [x] T010 Add `src/web/src/components/coin/__tests__/CoinViewer3D.test.ts` for two-sided flip, single-image disabled flip, missing image placeholder, open-image event, and reduced-motion class.

## Phase 3: Swipe Gallery Integration

- [x] T011 Update `src/web/src/components/SwipeGallery.vue` to use `<CoinViewer3D>` for the active card image area.
- [x] T012 Remove the current emoji flip button and scaleX flip animation from `SwipeGallery.vue`.
- [x] T013 Preserve swipe drag, next/previous navigation, page-change behavior, and tap-to-detail outside the flip control.
- [x] T014 Update `src/web/src/components/__tests__/SwipeGallery.test.ts` to prove the flip control exists and clicking it does not route to detail while existing pagination tests still pass.

## Phase 4: Coin Detail Hero Integration

- [x] T015 Update `src/web/src/pages/CoinDetailPage.vue` to use a hero-sized `<CoinViewer3D>` when either obverse or reverse image exists.
- [x] T016 Preserve existing `ImageLightbox` behavior by opening the current face image from the viewer's `open-image` event.
- [x] T017 Preserve wishlist purchase CTA, sell/delete behavior, metadata sections, and missing-image placeholder behavior.
- [x] T018 Add or update `src/web/src/pages/__tests__/CoinDetailPage.test.ts` for hero viewer rendering and open-image behavior.

## Phase 5: Documentation and Validation

- [x] T019 Update `specs/222-3d-coin-viewer/spec.md` acceptance criteria to match the implemented behavior.
- [x] T020 Check off completed tasks in `specs/222-3d-coin-viewer/tasks.md`.
- [x] T021 Run targeted tests from `src/web`: `npm.cmd test -- CoinViewer3D SwipeGallery useDeviceOrientation CoinDetailPage --run`.
- [x] T022 Run full frontend tests: `npm.cmd test`.
- [x] T023 Run `npm.cmd run type-check`.
- [x] T024 Run `npm.cmd run build`.
- [x] T025 Confirm the final diff contains no API/backend changes and no generated `dist/` artifacts.

## Dependencies and Execution Order

1. Phase 1 helpers support the component but can be implemented in parallel with static component tests.
2. Phase 2 blocks both Swipe Gallery and Coin Detail integration.
3. Phase 3 and Phase 4 can proceed independently after `CoinViewer3D.vue` exists.
4. Phase 5 is final validation and documentation.

## Implementation Notes

- Keep v1 frontend-only and do not add WebGL/three.js.
- Use a dedicated flip control in swipe gallery so card tap still opens detail.
- Coins with one image render as a static disc and keep the flip button disabled.
- Gyroscope permission must be requested only from a user gesture and denial must leave flip working.
- Respect `prefers-reduced-motion` for both flip animation and tilt.
