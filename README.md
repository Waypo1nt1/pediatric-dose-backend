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

| Метод | Адрес | Что делает |
| --- | --- | --- |
| GET | `/api/drugs` | список опубликованных препаратов, фильтр `min_adult_dose` и `max_adult_dose` |
| GET | `/api/drugs/feed` | первый опубликованный препарат |
| GET | `/api/drugs/feed/:id` | препарат по ид, с `?next=true` — следующий за ним |
| GET | `/api/drugs/draft` | черновик текущего пользователя |
| POST | `/api/drugs` | создать черновик: поле `drug_name`, файлы `image` и `video` |
| PUT | `/api/drugs/publish` | опубликовать черновик: `short_info`, `recommended_adult_dose_mg`, `max_daily_dose_mg` |
| DELETE | `/api/drugs/:id` | удалить свой препарат, статус меняется на `deleted` |
| POST | `/api/drugs/:id/like` | лайк, поле `value`: 1 ставит, 0 убирает |
| POST | `/api/users/register` | регистрация: `login`, `password` |
| POST | `/api/users/login` | аутентификация, заглушка до четвёртой лабораторной |
| POST | `/api/users/logout` | деавторизация, заглушка до четвёртой лабораторной |

Препарат возвращается одним и тем же набором полей: `id`, `drug_name`, `short_info`, `image_url`,
`video_url`, `recommended_adult_dose_mg`, `max_daily_dose_mg`, `likes_count`, `is_creator`.
Признак `is_creator` равен 1, если препарат создал текущий пользователь.

Файлы при добавлении уходят в MinIO, в бакет `drug-media`. Имя объекта собирается из ид препарата:
`drug_11.jpg` и `drug_11.mp4`. В таблицу записываются полные адреса файлов.

## Таблицы

`users` — пользователи: `id`, `login`, `password`, `is_moderator`.

`drugs` — препараты: `id`, `drug_name`, `short_info`, `drug_status`, `image_url`, `video_url`,
`recommended_adult_dose_mg`, `max_daily_dose_mg`, `created_at`, `creator_id`, `published_at`.
Статус: `draft`, `published`, `deleted`. Внешний ключ `creator_id` на `users`.

`drug_likes` — лайки, связь многие-ко-многим: `id`, `user_id`, `drug_id`. Внешние ключи на обе таблицы,
пара `user_id` и `drug_id` уникальна.

Каскадного удаления нет, все внешние ключи с `RESTRICT`. Больше одного черновика у пользователя быть
не может, за этим следит частичный индекс `idx_drugs_single_draft`.

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
