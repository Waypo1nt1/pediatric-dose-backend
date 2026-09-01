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

Go 1.25, gin, стандартный шаблонизатор `html/template`, logrus, MinIO для хранения изображений и видео.

## Страницы и методы

| Страница | HTTP метод | Контроллер |
| --- | --- | --- |
| Лента препаратов | `GET /drug-feed/1?next=true` | `Handler.GetDrugFeed` |
| Добавление препарата | `GET /drug-draft` | `Handler.GetDrugDraft` |
| Каталог препаратов | `GET /drugs?max_adult_dose=500` | `Handler.GetDrugCatalog` |

Фильтрация каталога выполняется на сервере по рекомендованной дозе для взрослого. Количество лайков вычисляется в контроллере по коллекции. JavaScript в приложении не используется.

Статусы препарата: `draft`, `published`, `deleted`. Черновик отображается только на странице добавления, удалённый препарат не отображается нигде.

## Структура

```
cmd/pediatric-dose-backend   точка входа
internal/api                 конфигурация сервера и маршруты
internal/app/repository      коллекция препаратов и доступ к данным
internal/app/handler         контроллеры страниц
templates                    шаблоны трёх страниц и панели вкладок
resources/styles             таблица стилей
```

## Запуск

```
docker compose up -d
go run ./cmd/pediatric-dose-backend
```

Приложение поднимается на `http://localhost:8080`. Базовый адрес хранилища переопределяется переменной окружения `DRUG_MEDIA_BASE_URL`, по умолчанию `http://localhost:9000/drug-media`.

## Хранилище MinIO

Изображения и видео препаратов хранятся в MinIO, в модели указаны только ключи файлов на латинице. Ключи заданы в `internal/app/repository/repository.go` и должны совпадать с именами объектов в бакете точно.

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
