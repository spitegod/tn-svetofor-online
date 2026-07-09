# Git Workflow

Проект ведется через `main` и короткоживущие рабочие ветки.

## Основные ветки

- `main` — стабильная ветка с актуальным рабочим кодом.
- `feature/*` — новые функции.
- `fix/*` — исправления ошибок.
- `chore/*` — технические задачи без изменения поведения.
- `docs/*` — документация.

## Как работать с задачей

Перед началом работы обновить `main`:

```bash
git checkout main
git pull origin main
```

Создать ветку от `main`:

```bash
git checkout -b feature/task-name
```

Примеры названий:

```text
feature/auth-page
feature/snapshots-list
fix/login-validation
chore/project-init
docs/update-readme
```

## Коммиты

Формат коммита:

```text
type: краткое описание
```

Типы:

- `feat` — новая функциональность.
- `fix` — исправление ошибки.
- `chore` — техническое изменение.
- `docs` — документация.
- `refactor` — рефакторинг без изменения поведения.
- `style` — стили и верстка.
- `test` — тесты.

Примеры:

```bash
git commit -m "feat: add login page"
git commit -m "fix: handle empty snapshots list"
git commit -m "docs: add git workflow"
```

## Pull Request

Рабочий процесс:

1. Сделать задачу в отдельной ветке.
2. Проверить запуск и сборку.
3. Запушить ветку.
4. Открыть Pull Request в `main`.
5. После проверки смержить PR.

```bash
git push origin feature/task-name
```

## Правила

- Не коммитить `.env`, локальные базы, логи и `node_modules`.
- Не работать напрямую в `main`, кроме совсем мелких правок по договоренности.
- Перед началом новой задачи всегда подтягивать свежий `main`.
- Один PR — одна логически завершенная задача.
- Если задача большая, лучше разбить ее на несколько маленьких PR.

## Проверки перед push

Backend:

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend
npm run build
```
