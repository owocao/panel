# AGENTS.md

This file is the working contract for AI coding agents in `/project/panel`.

## Project

- Name: `biu-panel`.
- Goal: personal lightweight web navigation panel plus bookmark manager.
- Reference: Sun-Panel user experience, but do not clone it exactly.
- Primary user: one personal administrator.
- Deployment target: Docker for final release, but local non-Docker development before v1.0.
- Default HTTP port: `55088`.
- HTTPS is handled by reverse proxy, not by the app.

## Confirmed Stack

- Backend: Go.
- Frontend: Vue 3 + Vite.
- Database: SQLite.
- Frontend package manager: npm.
- Final delivery: Dockerfile plus `docker-compose.yml`.
- Development: run backend and frontend directly on the host; do not force Docker for normal coding/debugging.

## Repository Layout

Use this structure unless there is a strong reason to change it:

```text
/project/panel/
  AGENTS.md
  docs/
  backend/
  frontend/
  deploy/
```

Documentation belongs in `/project/panel/docs`. Important docs should also be mirrored to Siyuan Notes following the user's existing project-document habit.

## Product Scope V1

V1 includes:

- Single administrator login.
- First-run initialization page if no admin exists.
- Initial admin via Docker environment variables for deployment.
- Sun-Panel-like homepage: group title plus small grid cards.
- Homepage groups with collapse/expand and drag sorting.
- Homepage cards with name, icon, group, LAN URL, WAN URL, sort order.
- LAN/WAN switch with global manual mode, auto-detect mode, timeout, unified detect URL, and per-card override.
- Left drawer bookmark manager, lazy-loaded only when opened.
- Infinite-level bookmark folders.
- Bookmark URLs with title, URL, favicon, plain-text note, sort order.
- Bookmark fuzzy search only; no homepage search.
- Bookmark folder and URL drag sorting, including cross-folder moves.
- Browser bookmark import/export compatible with Chrome and Safari HTML bookmark files.
- Manual `.tar.gz` system backup and restore.
- S3-compatible storage for uploads and backups.
- Local uploads for icons/logo/background.
- Site title, logo, background image/color settings.
- Mobile support.

V1 explicitly excludes:

- Multi-user isolation.
- Docker container management.
- Server monitoring.
- Homepage search.
- Homepage site online checks.
- Recycle bin.
- Captcha.
- Complex permission system.
- Sub-path deployment.
- Markdown notes.
- Tags.

## UX Rules

- Homepage must stay fast: do not load bookmark tree or bookmark data on initial homepage load.
- Bookmark drawer opens from the left and can be closed.
- Bookmark drawer layout: folder tree on the left, current folder URL list on the right.
- Bookmark tree loads on demand; expanding a folder loads its children.
- Homepage cards open in the current tab by default.
- Card edit operations are in a right-click menu, following Sun-Panel style.
- Mobile uses long-press to open the action menu.
- Bookmark URL menu includes: open, open in new tab, open in new window, edit, delete, copy link, move to folder, set/unset homepage card, batch selection entry.
- Deletion requires a second confirmation and is permanent.
- Drag sort saves immediately.

## Bookmark Import/Export Rules

- Preserve browser root folders, e.g. bookmark bar and other bookmarks.
- During one import, do not deduplicate internally; preserve source structure, order, and duplicates.
- For later imports, compare against existing data only.
- Treat as duplicate only when URL, title, note/description, and key fields are the same under the same folder.
- Different folders may contain identical URLs and titles.
- Preserve original URLs as much as possible. Do not apply aggressive normalization.
- Browser bookmark export exports bookmarks only, not homepage cards.
- Homepage cards and settings are covered by system backup instead.

## Data And Storage Rules

- Use SQLite for local data.
- Final Docker volumes should be split:
  - `./data/db:/app/data/db`
  - `./data/uploads:/app/data/uploads`
  - `./data/backups:/app/data/backups`
- S3 config is edited in the web settings page and stored locally.
- S3 fields: endpoint, region, bucket, access key, secret key, path-style switch, upload prefix.
- Do not add extra encryption for S3 secret key in V1.
- Normal icons/logo uploads: common image formats only, default max 5 MB.
- Background images: image type required, size not limited in V1.

## Security Rules

- Login session expires when the browser closes by default.
- Support remember-login for long-lived login.
- Log only login success and login failure.
- Lock login for 15 minutes after 5 consecutive failures.
- No captcha in V1.
- Never commit secrets, tokens, local `.env` files, database files, uploaded files, or backups.

## Backup Rules

- System backup format: `.tar.gz`.
- Backup includes SQLite DB, local uploaded files, config, and version metadata.
- S3-hosted images are not included by default; this may be optional later.
- Restore overwrites current data after a second confirmation.
- Backup version mismatch warns the user but allows forced restore after confirmation.

## Coding Rules

- Keep code simple and maintainable; personal-use performance and reliability matter more than enterprise abstraction.
- Prefer clear module boundaries: auth, settings, navigation, bookmarks, storage, backup, import/export.
- Do not introduce heavy frameworks unless needed.
- Avoid global npm installs.
- Keep frontend dependencies project-local.
- Add tests for parser/import/export logic and critical backend handlers.
- Do not write generated artifacts or dependency folders into git.
- Use Chinese UI text by default; no i18n in V1.
- Use ASCII in code unless existing files or UI copy clearly require Chinese.

## Documentation Rules

- Main requirements doc: `docs/requirements-v0.1.md`.
- Keep docs updated when product decisions change.
- Mirror important project docs to Siyuan Notes.
- Based on observed Siyuan structure, project docs likely belong under notebook `AI-creation`, using a parent project page with versioned child documents, similar to the existing `测试机使用记录系统` project.
- Before writing to Siyuan, confirm the target parent page if uncertain.

## Current Decisions Snapshot

- Project directory: `/project/panel`.
- Project name: `biu-panel`.
- Use Go installed via Debian `apt`.
- Use npm, not pnpm for now.
- First step before coding: write requirements and project-agent rules.
