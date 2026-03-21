# Social Feature — Implementation Plan

## Problem Statement
Add social capabilities to the coin collection app: user following, public/private profiles, follower galleries, coin comments/ratings, avatar uploads, and PWA navigation changes.

## Clarified Requirements
- **Follow model**: One-way, no approval needed
- **Privacy**: Private by default; opt-in to public; individual coins can be marked private
- **Ratings**: 1-5 star rating + text comments on coins
- **Profile images**: Custom avatar upload (default: Ed-Mar coin logo)
- **Email**: Required for new registrations; existing users prompted on next login
- **Follower detail view**: Images, name, ruler, era, denomination, grade only
- **Follower gallery**: Grid view only (no swipe)
- **Desktop nav**: Followers page in bottom tab bar
- **PWA nav**: Followers in hamburger; logout to hamburger; Home removed; Ed-Mar = home

## Phase 1: Backend Models & Database
- 1.1 Extend User model (Email, AvatarPath, IsPublic, Bio)
- 1.2 Extend Coin model (IsPrivate)
- 1.3 Create Follow model
- 1.4 Create CoinComment model
- 1.5 Update database migration

## Phase 2: Backend API — User Profile & Avatar
- 2.1 Avatar upload/delete endpoints
- 2.2 Profile update endpoint
- 2.3 Public user profile endpoint
- 2.4 User search endpoint
- 2.5 Update auth/register to require email
- 2.6 Email-missing flag on /auth/me

## Phase 3: Backend API — Social
- 3.1 Follow/unfollow + list followers/following + follower coins
- 3.2 Follower coin detail endpoint (limited fields)
- 3.3 Comment/rating CRUD endpoints
- 3.4 Register routes in main.go

## Phase 4: Frontend Types & API Client
- 4.1 Extend TypeScript types
- 4.2 Add API client functions

## Phase 5: Frontend — Settings & Profile
- 5.1 Avatar upload in Settings
- 5.2 Privacy/social settings
- 5.3 Email prompt modal
- 5.4 Registration page email field
- 5.5 Per-coin privacy toggle

## Phase 6: Frontend — Followers Pages
- 6.1 FollowersPage (tabs + search modal)
- 6.2 FollowerGalleryPage (grid-only, read-only)
- 6.3 FollowerCoinDetailPage (limited details + comments + ratings)

## Phase 7: Frontend — Navigation Changes
- 7.1 Desktop nav — add Followers link
- 7.2 PWA hamburger — add Followers + Logout
- 7.3 PWA App.vue — Ed-Mar = home, remove Home icon + top-bar logout
- 7.4 Router — add new routes

## Phase 8: Integration & Testing
- 8.1 Build verification
- 8.2 Manual testing checklist
