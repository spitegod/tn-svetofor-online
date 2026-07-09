# TN

Новый проект на стеке Vue 3 + TypeScript и Go.

## Структура

- `frontend` — клиентское приложение на Vue 3.
- `backend` — Go API.
- `docs` — документация проекта.

## Локальный запуск

Backend:

```bash
cd backend
go run ./cmd/server
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

По умолчанию:

- frontend: `http://127.0.0.1:5173`
- backend: `http://127.0.0.1:8080`
- health: `http://127.0.0.1:8080/api/health`
