# Расчёт дозы лекарства по площади поверхности тела

Бэкенд заявочной системы для расчёта индивидуальной детской дозы препарата по формуле Mosteller.

Курс «Проектирование систем и продуктовая веб-разработка», РИП 2026, МГТУ им. Н. Э. Баумана. Раздел «Медицина», вариант 17.

## Предметная область

- **Услуга** – препарат с рекомендованной дозировкой для взрослого.
- **Заявка** – расчёт индивидуальной детской дозы на основе роста и веса пациента.
- **Создатель заявки** – врач, **модератор** – клинический фармаколог.

Площадь поверхности тела по формуле Mosteller:

```
BSA (м²) = sqrt( рост_см * вес_кг / 3600 )
детская доза (мг) = взрослая доза (мг) * BSA / 1.73
```

Поля по предметной области:

| Таблица | Поля по предметной области |
| --- | --- |
| Препараты (услуги) | рекомендованная доза для взрослого, максимальная суточная доза |
| Расчёт дозы (заявка) | рост пациента, вес пациента |
| Позиция расчёта (м-м) | комментарий врача для ручного ввода, рассчитанная детская доза заполняется автоматически при формировании заявки |

## Стек

Go 1.25, gin, стандартный шаблонизатор `html/template`, logrus, PostgreSQL 17, gORM, Adminer, MinIO для хранения изображений и видео.

## Страницы и методы

| Страница | HTTP метод | Контроллер | Доступ к БД |
| --- | --- | --- | --- |
| Лента препаратов | `GET /drug-feed/1?next=true` | `Handler.GetDrugFeed` | gORM |
| Добавление препарата | `GET /drug-draft` | `Handler.GetDrugDraft` | gORM |
| Каталог препаратов | `GET /drugs?min_adult_dose=0&max_adult_dose=500` | `Handler.GetDrugCatalog` | gORM |
| Добавление, кнопка «Далее» | `POST /drug-draft` | `Handler.CreateDrugDraft` | gORM |
| Добавление, кнопка «Опубликовать» | `POST /drug-draft/publish` | `Handler.PublishDrugDraft` | gORM |
| Каталог, кнопка удаления | `POST /drugs/delete` | `Handler.DeleteDrug` | SQL `UPDATE` без ORM |

Фильтрация каталога выполняется на сервере по рекомендованной дозе для взрослого. Количество лайков считается запросом к таблице `drug_likes`. JavaScript в приложении не используется.

Статусы препарата: `draft`, `published`, `deleted`. Удаление логическое: статус меняется на `deleted`, такой препарат не отображается нигде и по его адресу возвращается 404.

Создатель зафиксирован константой, пользователь с `id = 1`. Если у него нет черновика, страница добавления просит наименование, изображение и видео, а кнопка «Далее» создаёт черновик. Если черновик есть, страница открывается с заполненными полями, и после краткого описания и обеих доз кнопка «Опубликовать» публикует препарат. Больше одного черновика у пользователя быть не может, это дополнительно проверяет уникальный частичный индекс `idx_drugs_single_draft`.

## Изображения и видео по умолчанию

Новые изображения и видео не сохраняются в БД и не передаются на сервер: у полей выбора файла нет атрибута `name`. Файлы по умолчанию лежат на сервере вместе со стилями и шрифтами:

```
resources/media/default-drug.jpg
resources/media/default-drug.mp4
```

Если `image_url` или `video_url` пустые, шаблон сразу подставляет файлы по умолчанию. Если адрес указан, но файл недоступен, браузер показывает запасное содержимое тега `<object>` и переходит ко второму `<source>` у видео.

## База данных

| Таблица | Назначение | Ключи |
| --- | --- | --- |
| `users` | пользователи: врачи и клинический фармаколог | PK `id` |
| `drugs` | препараты (услуги) | PK `id`, FK `creator_id` → `users.id` |
| `drug_likes` | лайки, связь м-м пользователь – препарат | PK `id`, FK `user_id` → `users.id`, FK `drug_id` → `drugs.id` |

Внешние ключи созданы с `ON DELETE RESTRICT`, каскадного удаления нет. Модели описаны в `internal/app/ds`, таблицы создаются миграцией `cmd/migrate`.

## Структура

```
cmd/pediatric-dose-backend   точка входа
cmd/migrate                  миграция моделей в PostgreSQL
internal/api                 конфигурация сервера и маршруты
internal/app/ds              модели таблиц
internal/app/dsn             строка подключения из переменных окружения
internal/app/repository      доступ к данным через gORM и SQL
internal/app/handler         контроллеры страниц
templates                    шаблоны трёх страниц и панели вкладок
resources/styles             таблица стилей
resources/fonts              шрифт Inter для `@font-face`
resources/media              изображение и видео по умолчанию
```

## Запуск

В корне проекта создаётся файл `.env`, он не хранится в git:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=pediatric_dose_user
DB_PASS=pediatric_dose_password
DB_NAME=pediatric_dose
```

Эти же переменные использует `docker-compose.yml` для контейнера PostgreSQL.

```
docker compose up -d
go run ./cmd/migrate
go run ./cmd/pediatric-dose-backend
```

После миграции таблицы заполняются через Adminer: `http://localhost:8081`, система PostgreSQL, сервер `postgres`, пользователь, пароль и база из `.env`.

Приложение поднимается на `http://localhost:8080`.

## Хранилище MinIO

Изображения и видео препаратов хранятся в MinIO, в таблице `drugs` указаны полные адреса в полях `image_url` и `video_url`, например `http://localhost:9000/drug-media/paracetamol.jpg`. Имена объектов в бакете записаны латиницей.

| Препарат | Изображение | Видео |
| --- | --- | --- |
| Парацетамол | `paracetamol.jpg` | `paracetamol.mp4` |
| Ибупрофен | `ibuprofen.jpg` | `ibuprofen.mp4` |
| Амоксициллин | `amoxicillin.jpg` | `amoxicillin.mp4` |
| Цефтриаксон | `ceftriaxone.jpg` | `ceftriaxone.mp4` |
| Метотрексат | `methotrexate.jpg` | `methotrexate.mp4` |
| Ципрофлоксацин | `ciprofloxacin.jpg` | `ciprofloxacin.mp4` |
| Омепразол | `omeprazole.jpg` | `omeprazole.mp4` |
| Диклофенак | `diclofenac.jpg` | `diclofenac.mp4` |
| Азитромицин | `azithromycin.jpg` | `azithromycin.mp4` |
| Дексаметазон | `dexamethasone.jpg` | `dexamethasone.mp4` |

Создание бакета:

```
docker exec -it minio_storage mc alias set myminio http://localhost:9000 root rootpassword
docker exec -it minio_storage mc mb myminio/drug-media
docker exec -it minio_storage mc anonymous set public myminio/drug-media
```

Загрузка файлов из локальной папки `media`:

```
docker cp ./media/. minio_storage:/tmp/media
docker exec minio_storage sh -c "mc cp --attr 'Content-Type=image/jpeg' /tmp/media/*.jpg myminio/drug-media/"
docker exec minio_storage sh -c "mc cp --attr 'Content-Type=video/mp4' /tmp/media/*.mp4 myminio/drug-media/"
```

`Content-Type` указывается явно: иначе объекты отдаются как `application/octet-stream` и браузер не проигрывает видео, а предлагает скачать файл.

Консоль хранилища: `http://localhost:9001`, логин `root`, пароль `rootpassword`. Проверка: `http://localhost:9000/drug-media/paracetamol.jpg` открывается без авторизации.

Видео вертикальное, 720 × 1280, соотношение сторон 9:16.
