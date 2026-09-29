# Расчёт детской дозы препарата

Вторая лабораторная: три страницы на Go с шаблонизатором, данные в PostgreSQL.
Услуга — препарат с дозировкой для взрослого, тема 17, раздел «Медицина».

## Запуск

В корне нужен файл `.env`:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=pediatric_dose_user
DB_PASS=pediatric_dose_password
DB_NAME=pediatric_dose
```

```
docker compose up -d
go run ./cmd/migrate
go run ./cmd/pediatric-dose-backend
```

Приложение — http://localhost:8080, Adminer — http://localhost:8081, консоль MinIO — http://localhost:9001.
Таблицы наполняются через Adminer импортом SQL-файла.

## Страницы и методы

| Метод | Адрес | Что делает |
| --- | --- | --- |
| GET | `/drugs` | плитка препаратов, фильтр по дозе для взрослого |
| GET | `/drug-feed/1` | лента, с `?next=true` открывает следующий препарат |
| GET | `/drug-draft` | страница добавления |
| POST | `/drug-draft` | кнопка «Далее», создаёт черновик |
| POST | `/drug-draft/publish` | кнопка «Опубликовать» |
| POST | `/drugs/delete` | удаление препарата, SQL `UPDATE` без ORM |

Пять методов работают через gORM, удаление — обычным SQL-запросом. Удаление логическое: статус
меняется на `deleted`, такой препарат нигде не виден и по его адресу возвращается 404.
Лайки берутся из таблицы `drug_likes`, JavaScript не используется.

Создатель задан константой, это пользователь с `id = 1`. Черновик у пользователя может быть только
один, за этим следит частичный индекс `idx_drugs_single_draft`.

Если `image_url` или `video_url` пустые или файл недоступен, показываются изображение и видео
по умолчанию из `resources/media`.

## Таблицы

`users` — пользователи: `id`, `login`, `password`, `is_moderator`.

`drugs` — препараты: `id`, `drug_name`, `short_info`, `drug_status`, `image_url`, `video_url`,
`recommended_adult_dose_mg`, `max_daily_dose_mg`, `created_at`, `creator_id`, `published_at`.
Статус: `draft`, `published`, `deleted`. Внешний ключ `creator_id` на `users`.

`drug_likes` — лайки, связь многие-ко-многим: `id`, `user_id`, `drug_id`.

Каскадного удаления нет, все внешние ключи с `RESTRICT`. Модели лежат в `internal/app/ds`,
таблицы создаёт миграция `cmd/migrate`.

## Хранилище MinIO

Файлы препаратов лежат в бакете `drug-media`, в таблице хранятся полные адреса вида
`http://localhost:9000/drug-media/paracetamol.jpg`.

```
docker exec -it minio_storage mc alias set myminio http://localhost:9000 root rootpassword
docker exec -it minio_storage mc mb myminio/drug-media
docker exec -it minio_storage mc anonymous set public myminio/drug-media
docker cp ./media/. minio_storage:/tmp/media
docker exec minio_storage sh -c "mc cp --attr 'Content-Type=image/jpeg' /tmp/media/*.jpg myminio/drug-media/"
docker exec minio_storage sh -c "mc cp --attr 'Content-Type=video/mp4' /tmp/media/*.mp4 myminio/drug-media/"
```

`Content-Type` нужно указывать явно, иначе браузер не проигрывает видео, а предлагает его скачать.
