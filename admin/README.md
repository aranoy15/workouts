# Workouts Admin

React + Vite admin panel for the Workouts backend.

## Setup

```bash
cp .env.example .env
npm install
npm run dev
```

Dev server: `http://localhost:5174`

Default API URL: `http://localhost:8080/api` (`VITE_API_URL`).

## Features

- Admin login (JWT via `X-Auth-Token`)
- Exercises: list, create, edit (multiple videos, muscle group/level catalogs), delete
- Users: list, create, update (including `is_active`)

## Deploy (Yandex Object Storage)

Default bucket: `workouts-admin`

```bash
export VITE_API_URL=https://your-backend.example/api
npm run deploy
# or
./scripts/deploy-s3.sh
```

Optional:
```bash
./scripts/setup-cdn.sh
./scripts/fix-content-type.sh
```

Override bucket: `export YC_BUCKET_NAME=workouts-admin`
