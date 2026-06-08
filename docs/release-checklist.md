# Release Checklist

## Build

- [x] Backend tests: `cd backend && go test ./...`
- [x] Frontend build: `cd frontend && npm run build`
- [x] Docker image build: `docker build -t biu-panel:release .`
- [x] Docker container smoke test

## Smoke Test Coverage

- [x] Container starts with `/app/data` bind mount
- [x] `/api/health` returns OK
- [x] Static frontend is served by backend
- [x] Admin can be initialized from environment variables
- [x] Login succeeds
- [x] Navigation group can be created
- [x] Navigation data returns empty item arrays instead of null

## S3

- [x] S3 settings are persisted
- [x] S3 test endpoint is implemented
- [x] S3 backup upload is implemented
- [x] S3 asset upload fallback is implemented
- [x] hi168 S3 path-style upload was tested successfully with a corrected bucket name

## Notes

- The Docker daemon on this machine had broken registry mirrors configured. They were removed from `/etc/docker/daemon.json`, and the original file was backed up as `/etc/docker/daemon.json.bak-biu-panel-*`.
- Runtime container currently runs as root so bind-mounted `./data:/app/data` works without manual `chown`, which is simpler for personal self-hosted deployment.
