# Архитектура

Проект стартует как модульный монолит:

- Backend разделяется на delivery, service, repository, config и app.
- Frontend использует FSD-подобную структуру без лишнего дробления.

## Frontend

```text
src/
├── app/
│   ├── router/
│   ├── providers/
│   └── App.vue
├── pages/
├── widgets/
├── features/
├── entities/
├── shared/
│   ├── ui/
│   ├── api/
│   ├── lib/
│   ├── composables/
│   ├── types/
│   └── assets/
└── main.ts
```

## Backend

```text
backend/
├── cmd/server/
├── internal/
│   ├── app/
│   ├── config/
│   ├── delivery/http/
│   ├── service/
│   └── repository/
└── migrations/
```
