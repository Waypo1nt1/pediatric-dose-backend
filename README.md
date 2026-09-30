# Расчёт детской дозы препарата

Веб-сервис на Go. Услуга — препарат с дозировкой для взрослого, по ней считается детская доза
по формуле Mosteller. Тема 17, раздел «Медицина».

Шаблонов и страниц в этой лабораторной нет, только API. Проверяется через Postman.

## Запуск

В корне нужен файл `.env`:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=pediatric_dose_user
DB_PASS=pediatric_dose_password
DB_NAME=pediatric_dose
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=root
MINIO_SECRET_KEY=rootpassword
MINIO_BUCKET=drug-media
```

```
docker compose up -d
go run ./cmd/migrate
go run ./cmd/pediatric-dose-backend
```

Сервис — http://localhost:8080, Adminer — http://localhost:8081, консоль MinIO — http://localhost:9001.
Таблицы наполняются через Adminer импортом SQL-файла.

## Методы

Текущий пользователь задан константой в `internal/app/singleton`.

| Метод | Адрес | Описание |
| --- | --- | --- |
| GET | `/api/drugs` | список опубликованных препаратов, фильтр `min_adult_dose` и `max_adult_dose` |
| GET | `/api/drugs/feed` и `/api/drugs/feed/:id` | лента: без ид — первый опубликованный, с ид — указанный, с `?next=true` — следующий за ним |
| GET | `/api/drugs/draft` | черновик текущего пользователя |
| POST | `/api/drugs` | создать черновик: поле `drug_name`, файлы `image` и `video` |
| PUT | `/api/drugs/publish` | опубликовать черновик: `short_info`, `recommended_adult_dose_mg`, `max_daily_dose_mg` |
| DELETE | `/api/drugs/:id` | удалить свой препарат, статус меняется на `deleted` |
| POST | `/api/drugs/:id/like` | лайк, поле `value`: 1 ставит, 0 убирает |
| POST | `/api/users/register` | регистрация: `login`, `password` |
| POST | `/api/users/login` | аутентификация, заглушка до четвёртой лабораторной |
| POST | `/api/users/logout` | деавторизация, заглушка до четвёртой лабораторной |

## Таблицы

| Таблица | Поля |
| --- | --- |
| `users` | `id` bigint, `login` varchar(25) уникальный, `password` varchar(100), `is_moderator` boolean |
| `drugs` | `id` bigint, `drug_name` varchar(100), `short_info` varchar(500), `drug_status` varchar(15), `image_url` varchar(255), `video_url` varchar(255), `recommended_adult_dose_mg` numeric(8,2), `max_daily_dose_mg` numeric(8,2), `created_at` timestamptz, `creator_id` bigint, `published_at` timestamptz |
| `drug_likes` | `id` bigint, `user_id` bigint, `drug_id` bigint |

Статусы препарата: `draft`, `published`, `deleted`. Внешние ключи: `drugs.creator_id` и `drug_likes.user_id`
на `users`, `drug_likes.drug_id` на `drugs`, все с `RESTRICT`. Пара `user_id` и `drug_id` уникальна.

## Структура

```
cmd/migrate                  создание таблиц по моделям
cmd/pediatric-dose-backend   точка входа
internal/api                 маршруты
internal/app/ds              модели
internal/app/dsn             строка подключения
internal/app/handler         обработчики
internal/app/repository      работа с БД
internal/app/schemes         структуры запросов и ответов
internal/app/singleton       текущий пользователь
```
