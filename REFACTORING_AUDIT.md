# TLD: аудит и план рефакторинга

Дата: **1 октября 2026**. Базовая ревизия: `bd1ef70996abd20382757c02a84e3f5143609473`.

Документ сначала был подготовлен как аудит без изменения кода. После запроса владельца «выполним весь этот план» начата реализация. Разделы 1–10 фиксируют **исходное состояние на базовой ревизии**; фактический статус изменений и остаток работ приведены в разделе 11. Поэтому утверждения о текущих дефектах и номерах строк в разделах 1–10 нельзя читать как описание уже изменённого рабочего дерева.

Главный вывод: проекту нужны общие контракты идентичности данных, фоновых задач, результатов проверок и поведения таблиц. Простое разбиение `model.go` по файлам оставит существующие ошибки. Начинать стоит с корректности этих контрактов, затем извлекать компоненты и оптимизировать их.

## 1. Область и степень проверки

Репозиторий содержит 104 Go-файла: 78 производственных, 26 тестовых; соответственно 17 607 и 5 357 строк. Проверены производственные пути `cmd`, `app`, `projectsync`, `registry`, `storage`, `ui`. Для безопасности выполнены независимые обзоры исходников и архитектурных границ, затем проверка выводов и локальные воспроизведения. Тестовые исходники изучены выборочно; полный аудит зависимостей по актуальным CVE не выполнялся.

Основа выводов — текущий код. Наблюдений за реальными пользователями и продуктовой аналитики нет: ниже описаны **паттерны, доступные через интерфейс**, а не доказанная частота их использования.

Проверки:

- `env GOCACHE=/tmp/tld-go-build-cache go test ./...` — успешно. Первый запуск упёрся в запрет песочницы на локальные порты `httptest`; повтор вне неё прошёл.
- `go vet ./...` — без диагностик анализатора; сборка `go build -o /tmp/tld-audit-build ./cmd/tld` — успешно. Сборка повторена вне песочницы после предупреждения о записи служебного кэша Go.
- 11 изолированных проверок воспроизвели описанные ниже дефекты. Проверочные файлы созданы во временной копии; исходные файлы и тесты репозитория не менялись. Успех этих проверок означает **подтверждение текущего дефекта**, а не его исправление.
- Проведён небольшой benchmark рендера; результаты приведены в разделе 4.
- Настоящая пользовательская БД, реальные PAT, внешние GitLab/GitHub/registry и операции изменения их настроек не использовались. Полная TUI-сессия в реальном терминале, Windows/Linux и `-race` не проверялись. Опасные terminal sequences в терминал не выводились.

Обозначения доказательств: **В** — локально воспроизведено; **К** — подтверждено чтением пути в коде; **Р** — рекомендация или условный риск, требующий отдельной проверки среды. Приоритет **P1** — исправить до расширения функций; **P2** — следующий этап; **P3** — улучшение сопровождения. Это приоритет разработки, а не CVSS.

## 2. Архитектура и границы ответственности

Текущий поток: `cmd/tld` → интерактивный пароль → SQLCipher → Bubble Tea `model.Update/View` → сервисы синхронизации → HTTP API / локальные сканеры → SQLCipher-кэш → таблицы.

| Область | Текущее устройство | Что сохранить / изменить |
|---|---|---|
| `internal/app/model.go` | 3 709 строк: навигация, формы, CRUD, команды, преобразования и состояние всех экранов | Оставить корневую маршрутизацию; переносить сценарии в модели экранов и прикладные сервисы |
| `internal/ui/creator.go` | `Creator.Render` принимает 66 параметров; сначала строит Dashboard даже для другого экрана | Принимать модель представления активного экрана и рендерить только его |
| `internal/projectsync` | Парсеры, HTTP-клиенты, проверки, изменения GitLab, кэширование и subprocess в одном пакете | Разделить по реальным границам: providers, parsers, scanners, use cases; небольшие интерфейсы потребителей |
| `internal/storage` | Конкретные SQL-репозитории, миграции, общая БД, одна connection | Сохранить SQLCipher и транзакции; извлечь идентичность и доменные значения из SQL-типов |
| `internal/ui/components` | Уже есть `TableCell`, `FormField`, `Modal`, `Progress`, `TabbedPanel` | Расширить поведенческими контрактами; сейчас переиспользуется в основном оформление |
| `internal/ui/uikit` | Палитра, символы, keymap, общие типы | Оставить стиль и клавиши; бизнес-правила и сравнение версий вынести из rendering |

Приложение локальное, с одним оператором; сетевого сервера и многопользовательской авторизации в нём не обнаружено. Основные недоверенные входы — содержимое подключённых репозиториев и ответы API. PAT хранится в зашифрованной БД и маскируется в форме. SQL-параметры передаются отдельно, сканеры запускаются через фиксированные argv без shell. Эти уже существующие меры нужно сохранить.

## 3. Конкретные проблемы

### R01. Смена проекта останавливает обработку сообщений синхронизации

**P1 · корректность / архитектура · В.** `internal/app/model.go:824–850`, `1553–1646`, `2858–2871`.

Сценарий: запустить `r` для A и перейти на B. Обработчик `projectSyncMsg` отбрасывает сообщение A и возвращает `nil`, не планируя следующий `waitProjectSync`. Даже сообщение `done` теряется, `Running` остаётся `true`, новый запуск блокируется. Если оставшихся сообщений достаточно для заполнения канала, producer также может заблокироваться; это следствие кода, отдельно не измерялось.

**Изменить:** отделить жизненный цикл задачи от текущего выделения. Все сообщения активного `JobID` обрабатывать до завершения; данные панели показывать по её `ProjectID`.

**Приёмка:** переходы A → B → A и на другой экран не мешают завершению; повторный запуск доступен; ни один канал не остаётся без consumer.

### R02. Асинхронные ответы не имеют достаточной идентичности

**P1 · корректность / архитектура · В/К.** `internal/app/model.go:106–110`, `135–143`, `524–555`, `623–630`, `1790–1883`; `internal/app/vulnerabilities.go:16–18,32–84`.

Команды захватывают выбранный проект/политику/стек, но `projectDependenciesLoadedMsg`, `policyValuesLoadedMsg`, `dependencyViewLoadedMsg`, `vulnsLoadedMsg` не позволяют проверить полный контекст запроса. Поздний ответ A способен заменить детали уже выбранного B. Для зависимостей это воспроизведено напрямую. `policyValueLatestLoadedMsg` аналогично может вписать старый ответ в другую или уже закрытую форму. Старый ответ списка способен перезаписать более свежий.

**Изменить:** `RequestID`/generation плюс `EntityID`, режим, период и ревизия входных данных; применять ответ только к тому состоянию, которое его запросило. Сброс/индикация загрузки деталей при смене master selection.

**Приёмка:** задержанные ответы в обратном порядке, закрытие формы и смена prod/dev не подменяют текущие данные.

### R03. Кэш смешивает разные источники одного типа

**P1 · целостность / безопасность · В.** `internal/projectsync/service.go:114–154`; `internal/projectsync/checks.go:445–452`; `internal/projectsync/vulnerabilities.go:127–190`; `internal/projectsync/releases.go:100–112`; `internal/projectsync/release_status.go:129–144`.

Ключ содержит тип провайдера и его project ID, но не `Source.ID`/origin. GitLab A/project 42 и GitLab B/project 42 имеют один ключ отчёта. В изолированной SQLCipher-БД запись чистого результата B заменила для A результат с семью High. Для отчётов и checks совпадение SHA не требуется. Файловый кэш дополнительно использует короткий SHA.

**Изменить:** версионированный структурированный ключ: источник с учётом смены origin → устойчивый provider project ID → полный commit SHA, где нужен → вид данных/режим/версия схемы. Обновить **всех** readers, включая Dashboard, которые сейчас часто передают лишь `Source{Type: ...}`. Старые неоднозначные ключи инвалидировать.

**Приёмка:** два источника с одинаковыми ID не влияют друг на друга; изменение origin/project association не показывает старые данные. Security severity: **Low** из-за необходимой конфигурации нескольких источников и отсутствия автоматического привилегированного действия; приоритет исправления высокий из-за неверных решений пользователя.

### R04. «Не удалось проверить» превращается в «уязвимостей нет»

**P1 · безопасность / корректность · В.** `internal/projectsync/kotlin.go:29–32,151–160`; `vuln_osv_api.go:34–38,117–125`; `vuln_osv.go:44–46`; `vuln_trivy.go:33–40`; `internal/projectsync/vulnerabilities.go:113–120`.

Gradle map notation `implementation group: 'org.example', name: 'library', version: '1.0.0'` не распознаётся текущим парсером. Получается ноль OSV queries, но `Scanned=true`. CLI fallback не запускается; без подходящего lockfile Trivy тоже не добавит проверку. Этот путь воспроизведён без сети. Пользователь видит нули при непроверенных зависимостях. Security severity: **Medium** — автор репозитория может сохранить зависимость в сборке, скрыв её от обычного сценария проверки.

Смежный подтверждённый дефект: `parseNpmAudit` принимает JSON вида `{"error":{"code":"ENOLOCK"}}` как успешный пустой отчёт (`vuln_javascript.go:74–124`). Не любой реальный `npm audit` с ENOLOCK даст такой результат: `CombinedOutput` может включать текст stderr и привести к ошибке JSON. Реальная достижимость зависит от версии и конфигурации logging; проверен именно контракт парсера.

**Изменить:** результат содержит outcome, coverage, обнаруженные/проверенные/пропущенные пакеты, warnings, scanner/version и timestamp. Различать `success`, `partial`, `unsupported`, `failed`, `canceled`. Fallback выбирать также по полноте покрытия. Проверять схему JSON и явные error-поля, разделить stdout/stderr и exit status.

**Приёмка:** пустой проект, неподдерживаемый формат, нулевые findings и сбой сканера дают разные состояния; partial/error не перезаписывает успешный snapshot как «чистый».

### R05. Незакрытый Gradle-блок вызывает panic

**P1 · надёжность / безопасность · В.** `internal/projectsync/kotlin.go:173–188`; `service.go:80–86`.

Файл, заканчивающийся ровно `dependencies {`, приводит к срезу `text[len(text):len(text)-1]`. Panic воспроизведён с `recover` в тесте. Bubble Tea обрабатывает panic команды завершением программы; это не ошибка одной строки проекта. Файл кэшируется до парсинга, поэтому повтор может встретить тот же вход.

**Изменить:** проверять завершённость блока и границы; возвращать диагностируемую ошибку. Не исполнять Gradle для анализа произвольного репозитория. Добавить malformed-input cases и fuzzing парсера.

**Приёмка:** EOF, вложенность, скобки в строках и комментариях не завершают TUI. Security severity: **Low** — отказ локального приложения без доказанной потери данных.

### R06. Удалённый текст может управлять терминалом

**P1 · безопасность · В/К.** `internal/projectsync/javascript.go:59–76,101–125`; `internal/app/model.go:2320–2330`; `internal/ui/screens/projects_screen.go:243–251`; `internal/ui/components/table.go:15–27`.

Имя/версия зависимости проходят из manifest в SQL и `style.Render` без удаления управляющих последовательностей. Изолированный тест подтвердил сохранение ESC sequence в готовой строке. Возможны подмена отображения, а при разрешённом OSC 52 — изменение clipboard. RCE этим не доказано; содержимое не выводилось в живой терминал.

**Изменить:** единая граница `PlainText` → безопасный текст → доверенные стили. Нейтрализовать ESC/C0/C1 и отдельно определить обработку переносов. Применить к dependency metadata, advisory text и внешним ошибкам. Не удалять ANSI из всего итогового view: собственные стили приложения легитимны.

**Приёмка:** внешние управляющие символы остаются видимыми данными либо удаляются; кириллица, emoji и собственные цвета работают. Security severity: **Low**, с зависимостью части эффектов от политики терминала.

### R07. Выбранная строка уходит за границу большинства таблиц

**P1 · usability / performance · В/К.** `internal/ui/screens/projects_screen.go:148–171`; `settings_screen.go:45–63`; `releases_screen.go:48–69`; `policies_screen.go:100–125,367–374`; аналогично Stacks, Sources, Dependencies, Dependency View и список проектов Vulnerabilities.

Экраны строят строки всех элементов, затем обрезают первые `height`. Выделение продолжает двигаться ниже видимого диапазона. Для Projects воспроизведено: выбран project 30, а на экране остаётся начало списка. Dashboard и панель зависимостей проекта уже содержат собственные функции видимого окна — готовый локальный паттерн для обобщения.

**Изменить:** общий `TableState` с ID выделения, viewport offset и `EnsureVisible`; сначала вычислять видимый диапазон, затем рендерить его. Добавить PageUp/PageDown и переход к началу/концу списка.

**Приёмка:** выделенная строка видна при навигации, сортировке, фильтрации, удалении и resize; контракт один для всех таблиц.

### R08. После прокрутки CVE теряется выделение и описание

**P2 · usability · В.** `internal/ui/screens/vulnerabilities_screen.go:170–204`.

`visibleVulnItems` возвращает срез, а цикл сравнивает локальный `index` с глобальным `selectedIndex`. Когда окно начинается не с нуля, подсветка и описание относятся не к той позиции либо исчезают. Воспроизведено для выбранного элемента 7 и трёх видимых строк.

**Изменить:** возвращать start index или строки со стабильным ID; учесть высоту раскрытого описания в viewport.

**Приёмка:** первая, средняя и последняя CVE имеют правильную подсветку и описание при любой прокрутке.

### R09. Пробел в названии проекта поглощается обработчиком checkbox

**P2 · usability · В.** `internal/app/model.go:1117–1165,3245–3339,3702–3709`.

Общий `case space` срабатывает вне зависимости от фокуса, но меняет только Freezing/EndOfLife. В текстовом поле пробел теряется. Воспроизведено для `Name`. В целом ввод реализован ручным append/backspace; полноценного положения курсора и редактирования внутри строки нет.

**Изменить:** сначала отдавать клавишу активному типу поля; checkbox обрабатывает Space только при своём фокусе. Общий text input с курсором, перемещением и Unicode-aware удалением; picker и checkbox — отдельные типы.

**Приёмка:** можно набрать имя с пробелами, исправить середину URL, вставить Unicode и переключить checkbox без влияния на другое поле.

### R10. Повторный Enter запускает повторное сохранение

**P1 · корректность / usability · В.** `internal/app/model.go:964–1036`, save handlers `1895–2046`.

Форма остаётся открытой и доступной до ответа. Два Enter создают две команды save; это воспроизведено. Уникальные ограничения БД ограничивают часть дублей, но второй ответ может принести конфликт, закрыть новое состояние формы или изменить уже обновлённую запись.

**Изменить:** `FormSessionID`, `Submitting`, dirty state; одна команда на отправку, ошибки возвращать только инициировавшей форме. При отмене явно определить, что происходит с уже отправленной mutation.

**Приёмка:** двойной Enter даёт одну mutation; старый ответ не закрывает новую форму; после ошибки ввод и фокус сохраняются.

### R11. Часть ошибок вообще не видна пользователю

**P1 · usability / диагностируемость · К.** `internal/app/model.go:496–555,617–622,717–721,857–932`; `internal/app/dashboard.go:100–148`.

Ошибки чтения списков, кэша и проверки прав записываются в `m.err`, но `View` не передаёт это поле renderer. Успешное параллельное сообщение может затем присвоить ему `nil`. Отдельные формы/progress показывают свои ошибки, поэтому поведение зависит от места сбоя.

**Изменить:** явные `LoadState` и `ErrorState` на экран/панель; общий status surface и журнал последних операций. Хранить last-good data с отметкой stale. Сообщать источник, проект, действие и возможность retry, не раскрывая токен.

**Приёмка:** ошибки startup, 403, timeout, повреждённого кэша и SQL видны в соответствующем контексте и не исчезают из-за несвязанной загрузки.

### R12. Нет сквозной отмены и общего бюджета фоновой задачи

**P1 · архитектура / надёжность · К.** `internal/app/app.go:12–29`; `model.go:1553–1746`; `internal/projectsync/vulnerabilities.go:52–54,114`; `vuln_javascript.go:64–67`; `vuln_osv.go:68–76`; `vuln_trivy.go:71–75`.

HTTP-клиенты имеют timeout отдельного запроса, но команды используют `context.Background()`, а интерфейс `VulnStrategy.Scan` вообще не принимает context. Subprocess не получает deadline задачи. Каналы отправляют сообщения без `select` на отмену. Можно запустить длительную работу на разных экранах, но единого обзора и отмены нет; lifecycle закрытия store не координируется с собственными workers.

**Изменить:** root context приложения → job context → provider/parser/scanner. Общий runner с cancel, deadline, завершением workers перед закрытием ресурсов. Отдельные ограничения HTTP, всего проекта, subprocess, размера body/output и страниц; один фиксированный timeout для всех этих разных операций не подходит.

**Приёмка:** отмена завершает запрос и subprocess, освобождает worker и состояние Running; выход при активной работе не оставляет заблокированных отправителей.

### R13. Массовое обновление порождает квадратичное перечитывание кэша

**P2 · performance · К.** `internal/app/model.go:753–754,783–784,818–819`; `internal/app/checks.go:36–87`; `internal/projectsync/checks.go:78–93,147–175`; `internal/app/vulnerabilities.go:32–84`; `internal/storage/store.go:57`.

После каждого проекта `step` заново загружает все строки. Для Settings каждый reload читает 14 checks каждого проекта отдельными запросами. За N шагов это около **14 × N²** cache reads, плюс финальный reload; при N=100 — около 140 000. Это расчёт по коду, не SQL-профиль. Одна connection сериализует SQL. Перемещение по проектам в Vulnerabilities также перечитывает и декодирует отчёты всех проектов ради одной панели.

**Изменить:** обновлять строку завершившегося проекта; batch cache read для initial load; отдельный loader деталей выбранного проекта; объединять частые progress updates. Увеличение пула SQLite не устраняет причину.

**Приёмка:** число чтений массового обновления растёт линейно с числом проектов/checks; смена выделения не перезагружает все отчёты.

### R14. Рендер выполняет работу для невидимых данных

**P2 · performance · В/К.** `internal/ui/screens/projects_screen.go:34–76,148–171`; `internal/ui/creator.go:126–134`; `internal/ui/components/table.go:15–38`.

Список целиком форматируется на каждом View; ширины нескольких столбцов вычисляются проходами по всем проектам. Общий Creator дополнительно строит Dashboard до определения нужного body. Замер фиксированного viewport показал линейный рост затрат — раздел 4.

**Изменить:** выбрать screen до rendering; visible rows first; кэшировать расчёт ширин на изменении данных/размера; не декодировать данные в View. Оптимизацию строковых builders и style construction делать по профилю после этого.

**Приёмка:** повторное движение по таблице на 5000 строк не форматирует 5000 строк; benchmark на одном viewport почти не зависит от общего размера после подготовки данных.

### R15. Batch работает последовательно и прекращается на первой ошибке

**P2 · performance / usability · К.** `internal/app/checks.go:112–170`; `internal/app/vulnerabilities.go:117–180`; `internal/app/releases.go:106–168`; `internal/app/model.go:1698–1746`.

Суммируются задержки всех проектов; один timeout/ошибка сканера останавливает оставшиеся. Пропуски не сохраняются как отдельный итог, завершение может сообщать `complete`, хотя часть проектов пропущена. В Releases `ReleaseService` и `ReleaseStatusService` отдельно получают tags (`internal/projectsync/releases.go:55`, `release_status.go:40`), увеличивая число запросов.

**Изменить:** ограниченная конкуренция с лимитом на источник, общий snapshot для связанных проверок, итог по каждому элементу: success/failed/skipped/canceled. Retry только неуспешных; для read-only запросов — ограниченная политика backoff, учитывающая rate limits. Автоматический retry mutations без проверки результата не вводить.

**Приёмка:** сбой B не скрывает результаты A/C; можно повторить только B; concurrency/число запросов проверены fake-клиентами.

### R16. Кэш не имеет политики свежести и размера

**P2 · performance / целостность · К.** `internal/storage/cache.go:68–139`; `internal/projectsync/service.go:130`; `internal/projectsync/vulnerabilities.go:161–174`; `internal/projectsync/checks.go:436`; `internal/projectsync/releases.go:107`.

Записи создаются с TTL=0. `DeleteExpired`/`ClearNamespace` есть, но в runtime-путях вызовов нет. Старые manifest snapshots накапливаются, отчёты не несут достаточной информации о commit, времени и версии scanner. Изменение advisory database может сделать старый чистый результат неактуальным даже без изменения проекта. `Force:true` в UI одновременно заставляет заново разбирать/сохранять зависимости неизменного commit.

**Изменить:** разделить immutable file cache и mutable result snapshots; freshness policy, byte/entry budgets, retention, schema version, invalidation при изменении источника/проекта. Обычный refresh и force rescan должны иметь разные явно описанные значения.

**Приёмка:** ограниченный размер кэша, проверяемый cleanup; видны commit/time/coverage и состояние stale; обычный повтор не переписывает идентичные зависимости без причины.

### R17. Чтение tags ограничено первой страницей

**P2 · корректность · К.** `internal/projectsync/source_clients.go:281–303,602–624,764–788`; `internal/projectsync/releases.go:44–71`; `release_status.go:99–105`.

GitLab, Gitea и Bitbucket запрашивают 100 tags и не продолжают pagination. Это кодовая причина неполного timeline/count и потенциально неверного отсутствия version tag. Для GitHub метод Tags пока возвращает unsupported (`source_clients.go:148–150`), а не полноценную поддержку Releases. Реальные ответы серверов в этой задаче не проверялись.

**Изменить:** provider-specific paginator с контролем origin, циклов, бюджета и полноты; capability у экрана. Передавать полную либо явно неполную выборку. Контракты даты tag/commit и compare timeout проверить отдельно на fixtures каждого провайдера.

**Приёмка:** >100 tags, пустая последняя страница, timeout в середине и unsupported provider не создают видимость полного успешного результата.

### R18. Сохранение Namespace не атомарно

**P2 · архитектура / целостность · К.** `internal/app/model.go:1915–1943`; `internal/storage/namespace.go:67–107`.

Create/Update основной записи и SetPolicy — отдельные команды SQL. При сбое второго шага первая часть уже сохранена, хотя форма сообщает ошибку. Предварительная проверка формы не заменяет транзакцию.

**Изменить:** один прикладной метод сохранения namespace вместе с policy в транзакции. Тот же принцип применить к целостному snapshot checks: сейчас отдельные checks пишутся по одному и сбой оставляет смесь запусков (`projectsync/checks.go:132–144`).

**Приёмка:** injected failure второго шага не оставляет половину обновления; failed check run не выглядит единым свежим snapshot.

### R19. Единая модель соединяет слишком много независимых состояний

**P2 · архитектура · К.** `internal/app/model.go:17–91,253–308,857–932`; `internal/ui/creator.go:57–123`; `internal/projectsync/types.go:69,73–107`.

Десятки selected IDs, отдельных форм, delete confirms, bool-флагов modal и статусов образуют неявную машину состояний. Приоритет модальных окон задан цепочкой `if`. Сервисы зависят от конкретных storage repositories, а `Dependency` является alias SQL-типа. Подмена I/O в сценарных тестах затруднена.

**Изменить:** `ScreenModel` на сценарий, одна активная modal state с session ID, общий job registry; доменные типы без UI/SQL, узкие интерфейсы cache/repository/provider/scanner. Сначала закрепить поведение, затем извлекать по одной вертикали.

**Приёмка:** новый экран не расширяет Render на десяток аргументов; сценарий можно тестировать с fake service; невозможны две одновременно активные формы одного взаимодействия.

### R20. Идентичность пакета и сравнение версий определяются эвристиками

**P2 · архитектура / корректность · К.** `internal/storage/dependency_view.go:253–267`; `internal/app/dashboard.go:242–302`; `internal/ui/screens/dependency_view_screen.go:180–253`; `internal/registry/registry.go:251–329`.

Нормализация удаляет `- _ . / : @` и приводит к lower case: `a-b`, `a/b`, `ab` становятся одной сущностью. Dashboard/View извлекают все числа из строки версии; prerelease, range и alias могут получить некорректный статус. В Registry существует третья реализация сравнения.

**Изменить:** `PackageRef{Ecosystem, CanonicalName}` и явные aliases; хранить declared constraint отдельно от resolved version. Единый domain API сравнения с адаптерами экосистем; `unknown/incomparable` — полноценный результат.

**Приёмка:** scoped npm names, Maven group/artifact, Go paths не схлопываются; ranges, prereleases и нестандартные версии не превращаются молча в точные числовые версии.

### R21. Dependency View теряет связь policy version с конкретным проектом

**P2 · архитектура / корректность · К.** `internal/storage/dependency_view.go:39–68,150–187`; `internal/ui/screens/dependency_view_screen.go:110–128`.

`viewColumns` выбирает DISTINCT dependency + policy version сразу по всем namespaces стека. Если политика A требует v1, а B — v2, появляются два столбца одной зависимости. Каждая строка получает actual по одному `DependencyID` и сравнивается с обеими policy versions, без указания принадлежности политики. Это способно показывать нерелевантное отклонение.

**Изменить:** expected version должна определяться policy проекта в ячейке. Если сравнение нескольких политик задумано как отдельная функция, явно выбрать режим и подписать policy/namespace.

**Приёмка:** два namespace с разными pins показывают корректное отклонение каждого проекта только от применимой политики.

### R22. Проверки смешивают отсутствие настройки, отсутствие данных и локальную политику

**P2 · архитектура / usability · К.** `internal/projectsync/checks.go:16–53,108–144,349–370,417–439`; `internal/projectsync/protection.go:27–35,69–92`.

Ошибки загрузки CI/protected branches частично сворачиваются в boolean «не найдено». Пользователь не отличает 403/timeout от отсутствующей защиты. Требования `master/dev`, конкретный owner nightly и Spectrum CI markers зашиты в код. `protect` проверяет уровни доступа, а операция создания защиты сохраняет уже существующие правила — она может успешно завершиться, не исправив этот check; документация это поведение подтверждает.

**Изменить:** `CheckResult{State, Reason, Evidence, ObservedAt, Remediation}` и настраиваемый policy profile поверх текущих defaults. Capability провайдера — отдельно. Операция должна объяснять, какие именно несоответствия исправляет; существующие protection rules не переписывать молча.

**Приёмка:** missing/denied/failed/unsupported различимы; детали check показывают ожидание и факт; профиль меняется без редактирования switch.

### R23. Подсказки и обработчики описывают разные возможности

**P2 · usability / архитектура · К.** `internal/ui/components/hints.go:114–151`; `internal/ui/uikit/keymap.go:34–62`; `internal/app/model.go:336–396,440–493`; `internal/app/nav_modal.go:8–32`; `internal/app/operations.go:37–55`.

Общие list hints показывают clone вне Projects, add/edit/delete на read-only View/Releases; Policies также показывает clone без обработчика. Tab переключает панель, стек или период в зависимости от экрана. Ctrl+C обрабатывается в формах, но не в navigation/operations modal. Поиск, фильтрация и быстрый выбор из длинного picker не реализованы.

**Изменить:** единый registry доступных actions по screen/focus/modal/job state; из него получать dispatch и hints. Зафиксировать роли Tab/Shift+Tab, стрелок, Esc, Ctrl+C. Сохранить установленные Settings bindings: `s` и `Shift+←/→`. Поиск — отдельный компонент списка/picker, а не копии обработчиков.

**Приёмка:** каждая показанная подсказка выполняет действие или объясняет disabled; скрытые unsupported actions не запускаются; Ctrl+C имеет единый контракт.

### R24. Макет не имеет явного режима для узкого терминала

**P2 · usability · К/Р.** `internal/ui/screens/settings_screen.go:13–15,66–73`; `internal/ui/components/table.go:15–31`; `modal.go:35–72`; `internal/ui/creator.go:289–311`.

Settings требует минимум 24 + 14×8 = 136 клеток только для столбцов. Ширина ячейки задаётся как style width, но общей политики обрезки/горизонтального viewport нет. Modal padding не ограничивает длинную строку. Logo + hints имеют свою минимальную ширину; шрифт предполагает специальные glyphs. Полная визуальная проверка матрицы терминалов пока не проведена.

**Изменить:** layout budget по display cells; sticky project column и horizontal viewport; compact/help overlay; ограниченная высота modal и прокрутка; fallback обозначения без Nerd Font. Конкретный минимальный supported viewport закрепить после визуальной проверки.

**Приёмка:** 80×24, 120×30, 160×50, resize, длинные URL/CVE, кириллица/emoji: доступ к действиям сохраняется, строки не ломают соседние панели.

### R25. Удалённые изменения нуждаются в явном контракте intent/result

**P2 · безопасность / usability · К/Р.** `internal/app/operations.go:37–100`; `internal/projectsync/operations.go:25–59`; `internal/projectsync/checks.go:184–193`; `internal/ui/components/operations_modal.go:16–37`.

Enter в меню сразу применяет выбранную операцию. Имя проекта уже показано — утверждать полное отсутствие контекста неверно, — но source origin, payload/разница и последствия не раскрыты. В частности, разделение CI cache намеренно отключается и считается здоровым состоянием. Это существующая политика продукта, а не автоматически доказанная уязвимость; её допустимость зависит от доверия к contributors/runners.

**Изменить:** для внешних mutations показывать точный target/source и изменение; объяснять security-sensitive последствия. Результат хранить по исходному target, выполнять readback и отдельно сообщать apply success / verification failure. Для многошаговой защиты показывать частичный итог. Дополнительное подтверждение нужно значимым внешним изменениям, а не каждому чтению или перемещению.

**Приёмка:** пользователь видит, к какому серверу/проекту относится действие и что изменится; повтор после timeout не дублирует уже выполненное вслепую.

### R26. Валидация URL и политика credential destination недостаточны

**P1 · security hardening / корректность · К/Р.** `internal/storage/source.go:105–120`; `internal/projectsync/source_clients.go:29–32,717–741,798–805,855–934`.

Проверяется непустой URL; schemeless/некорректный URL может превратиться в публичный endpoint провайдера. Например, введённый без scheme внутренний GitLab не означает фактический внутренний адрес запроса. Возможна ошибочная отправка PAT не тому сервису. У GitLab нет собственной redirect policy для `PRIVATE-TOKEN`; Bitbucket `next` становится новым аутентифицированным запросом без проверки origin.

**Изменить:** валидировать/canonicalize scheme, host, port до сохранения; показать effective API URL; публичный default выбирать явно. Привязать credentials к разрешённому origin, отклонять HTTPS downgrade и неожиданный переход; проверять pagination/cycles. Сохранить поддержку сознательно настроенных self-hosted/private адресов.

**Приёмка:** malformed URL отклоняется; чужой origin не получает PAT через redirect/next. Источник URL сейчас — локальный оператор, а контроль redirect/next внешним участником не доказан: это **не подтверждённые SSRF или удалённая кража токена**.

### R27. Глобальный индикатор Maintainer основан на одном проекте

**P2 · usability / модель прав · К.** `internal/app/dashboard.go:100–170`; `internal/projectsync/source_clients.go:412–435`.

Проверяется первый активный GitLab project, но UI сообщает обобщённое состояние токена. Право Maintainer относится к конкретному проекту/группе; при нескольких sources или разных membership индикатор вводит в заблуждение. Сервер всё равно проверяет mutation: обход авторизации этим не установлен.

**Изменить:** статус прав по source/project, timestamp и unknown/denied/error; UI capability конкретной операции не выводить из одного глобального boolean.

**Приёмка:** у двух проектов с разными правами независимые статусы; недоступный probe не означает отсутствие прав у всех.

## 4. Измеренные затраты рендера

Условия: macOS arm64, Apple M3 Pro, Go 1.27.1, `renderProjectsContent(110, 20, rows, 1)`, ASCII-названия, три итерации на размер, без сетевого I/O. Benchmark находился во временной копии: `go test ./internal/ui/screens -run '^$' -bench '^BenchmarkAuditProjectsRender$' -benchtime=3x -benchmem`.

| Проектов | Время одного рендера | Выделено памяти за рендер | Аллокаций |
|---:|---:|---:|---:|
| 100 | 1,43 мс | 1 884 576 B | 21 777 |
| 1 000 | 12,19 мс | 5 984 186 B | 216 185 |
| 5 000 | 57,88 мс | 32 763 642 B | 1 080 203 |

Это предварительный микрозамер, не p95 и не профиль всей TUI. Он подтверждает работу пропорционально всей коллекции при фиксированном окне. Первые оптимизации — viewport и отказ от неактивного rendering; целевую задержку и устойчивые baseline следует определить более длинным benchmark после изменения. В SQL и сети оценки из R13/R15 пока основаны на коде, а не на измеренных latency.

## 5. Общие правила взаимодействия и что привести к ним

| Правило | Единый контракт | Где применить |
|---|---|---|
| **U1. Выбрать → увидеть детали** | Stable ID, отдельный LoadState деталей, request generation, last-good data с отметкой stale | Projects → Dependencies, Policies → Values, Vulnerabilities → CVE, View → Stack; R02 |
| **U2. Перемещаться по списку** | `TableState`: selection, offset, visible range, sort/filter; выделение всегда видно | Все таблицы; переиспользовать подход Dashboard/Project dependencies; R07–R08 |
| **U3. Найти объект** | `/` открывает поиск; один reset/cancel; picker с фильтром; сохранить ID выбранного объекта | Projects, Sources, Dependencies, Policies, Settings, Releases, Vulnerabilities; формы выбора source/stack/dependency |
| **U4. Создать / изменить / клонировать** | Field types, focus order, dirty/submitting/session state, понятная ошибка поля | Stack/Namespace/Dependency/Project/Source/Policy/PolicyValue forms; R09–R10 |
| **U5. Удалить** | Target snapshot по ID, имя и последствия; один pending submit; ошибка оставляет контекст | Шесть delete-confirm состояний и удаление PolicyValue в `model.go`; общий confirm component |
| **U6. Обновить строку / группу** | `r` — текущий объект; `R` — явно описанный набор; один runner, cancel, per-item outcome и retry | Dependency sync, Checks, Releases, Vulnerabilities, Pin updates; R01/R12/R15 |
| **U7. Выполнить внешнее изменение** | Preview target/source/diff → apply → readback → частичный/полный итог | GitLab operations, в дальнейшем любые remote mutations; R25 |
| **U8. Понять состояние данных** | loading / empty / ready / stale / partial / error / unsupported; provenance/time | Все панели, особенно Dashboard, Checks, Releases и Vulnerabilities; R04/R11/R16/R22 |
| **U9. Узнать доступные клавиши** | Hints и dispatch из одного списка enabled actions с учётом фокуса/модального окна | `keymap.go`, `hints.go`, `model.go`, nav/operations modal; R23 |
| **U10. Переключить фокус / режим** | Tab/Shift+Tab — согласованная навигация; изменение mode/period явно подписано; Esc — ближайший контекст | Все двухпанельные экраны, View, Releases, формы; не менять существующие Settings `s`/Shift+стрелки без явной миграции |
| **U11. Прочитать внешний текст** | Raw input → safe plain text → styling; одна политика width/truncation/Unicode | Table cells, descriptions, status/error, picker и modal; R06/R24 |
| **U12. Работать с сущностью и snapshot** | Source/project/package/revision identity не зависит от подписи, позиции или форматирования | Cache, SQL view, async messages, policy comparison, scanner reports; R02/R03/R20/R21 |

Правила U1–U12 — предлагаемый целевой контракт. Поиск и часть навигации требуют добавления поведения, поэтому их следует включать в отдельные небольшие изменения с согласованными клавишами, а не маскировать под механическое перемещение кода.

### Инвентаризация экранов

| Экран / компонент | Что привести к общим правилам |
|---|---|
| Dashboard | U1/U8/U12: честные unknown/stale/partial, права с областью действия; существующий viewport взять за основу U2 |
| Stacks + Namespaces | U2/U4/U5/U9/U11; Namespace save — одна транзакция |
| Dependencies | U2/U3/U4/U5/U11/U12; registry/package identity не выводить из display name |
| Projects, обе панели | U1–U6/U8/U11; clone через тот же жизненный цикл формы, job не зависит от selection |
| Sources | U2–U5/U11/U12; effective endpoint, маскировка, secret reference вместо токена в display DTO |
| Policies, values, pin updates | U1–U6/U8/U12; latest-version response привязан к form session и package |
| Dependency View | U2/U3/U6/U8/U10/U12; горизонтальное окно и ожидаемая версия policy конкретного проекта |
| Settings | U2/U3/U6–U9/U11/U12; детали check, capability, версия snapshot, горизонтальное окно |
| Releases | U2/U3/U6/U8–U12; полнота pagination и корректная подпись period control |
| Vulnerabilities, обе панели | U1–U3/U6/U8–U12; coverage, scanner provenance, правильный selected index |
| Nav / Operations / Delete / Form modals | Один active modal, focus/session contract, Ctrl+C/Esc, width/height budget |

## 6. Целевая структура без смены технологического стека

Достаточно модульного Go-приложения с существующим Bubble Tea и SQLCipher. Предлагаемое разделение:

```text
internal/domain/          ProjectRef, PackageRef, Snapshot, CheckResult, ScanOutcome
internal/application/     sync/check/scan/release/policy use cases; narrow ports
internal/jobs/            context, JobID, progress, cancellation, bounded workers
internal/providers/       GitLab/GitHub/Gitea/Bitbucket/registry transport + capabilities
internal/parsers/         manifest readers returning data + coverage + diagnostics
internal/scanners/        npm/OSV/Trivy adapters and process runner
internal/storage/         SQL repositories, migrations, versioned cache adapter
internal/app/             Bubble Tea root routing and feature models
internal/ui/components/   table viewport, fields, picker, modal, status, safe text
internal/ui/screens/      presentation of screen view models
```

Это направление зависимостей, не требование немедленно создать все каталоги. Интерфейс вводится у потребителя: например, `ScanProject(ctx, request)` или чтение snapshot, а не универсальный CRUD-framework. Общими должны стать реально повторяющиеся правила; специфичные таблица timeline, матрица версий и детали CVE сохраняют свои renderers.

Минимальные доменные контракты:

- `ProjectRef`: source identity, provider ID, local ID; immutable snapshot по full SHA.
- `Job`: ID, kind, target set, status, progress, cancel, per-target result. Отдельно `RequestID` коротких UI-loads.
- `Snapshot`: schema version, source/project/revision, observed time, tool/policy version, outcome/coverage, данные и diagnostics.
- `Action`: ID, binding, visible/enabled/reason, handler; один источник для UI и dispatch.
- `TableState`: stable selected ID, visible offset, columns, sort/filter; данные хранятся отдельно.
- `FormState`: session, typed fields, focus, dirty, submitting, validation; одна mutation на session submit.

## 7. Последовательность изменений

| Этап | Содержание | Зависимости | Проверяемый результат |
|---|---|---|---|
| **1. Исправить доказанные нарушения контрактов** | R01–R06, R08–R11; минимальные regression tests; уточнить URL validation R26 | Не требует переезда пакетов | Данные не подменяются, задачи завершаются, parser/renderer безопасно обрабатывают вход, ошибки видны |
| **2. Упорядочить identity и snapshots** | R03/R04/R16/R18/R20/R21/R22; versioned cache и транзакции | Контракт результата из этапа 1 | Миграция старых ключей проверена; один snapshot имеет одну provenance; clean отличается от incomplete |
| **3. Извлечь job runner** | R12/R15; использовать из четырёх refresh workflows и pins; убрать `Background` внутри adapters | Job/result contracts | Cancel/timeout/partial/retry одинаковы; сначала сохранить текущую последовательность, затем добавить измеренную bounded concurrency |
| **4. Унифицировать списки и формы** | R07/R09/R10/R23/R24; U1–U11 и screen view models | Стабильные IDs и request generations | Табличные сценарии проходят на всех экранах, keyboard hints совпадают с обработкой |
| **5. Оптимизировать чтение и рендер** | R13/R14/R17; delta updates, batch cache reads, provider pagination, shared project snapshot | Новые viewport/result boundaries | Рост SQL-read count линейный; нет рендера невидимых строк; полные tags |
| **6. Завершить границы и эксплуатационную документацию** | R19/R25/R27; точные provider/scanner capabilities; README и диагностика | Извлечённые сценарии | Экран не знает SQL/credentials; документация описывает установку, supported formats, offline/error режимы |

Каждый этап делить по одной вертикали: сначала Projects как образец list/detail/form/job, затем Settings, Vulnerabilities и остальные. Не совмещать миграцию кэша, смену версионного алгоритма и перестройку всех экранов в одном большом diff. До миграций предусмотреть проверяемую резервную копию зашифрованной БД и тест восстановления на тестовом файле.

## 8. Набор приёмочных сценариев для последующих изменений

| Область | Обязательные случаи |
|---|---|
| Async | A → B до ответа A; обратный порядок двух ответов; смена stack/mode/period; закрытие/повторное открытие формы |
| Jobs | Выход и cancel; timeout; смена selection; повтор запуска; ошибка одного из N; skip; readback failure после успешной mutation |
| Tables | 0/1/1000 строк, начало/середина/конец, sort/filter/delete сохраняют ID, viewport после resize, multiline selected detail |
| Forms | Двойной Enter; Space в тексте и checkbox; paste/Unicode; SQL error; поздний latest-version response; отмена pending mutation |
| Cache | Два origin с project 42; смена origin; full SHA; stale snapshot; old schema; eviction; чтение после частичной ошибки |
| Scanners | Clean vs no coverage; malformed/error JSON; частичный отчёт; missing tool/lockfile; unsupported notation; subprocess timeout; stdout/stderr limit |
| Security boundaries | ESC/C0/C1 в metadata/errors; Gradle EOF/fuzz; redirect/next на другой origin; malformed URL; отсутствие секретов в diagnostics |
| Providers | Несколько страниц tags/protection; 401/403/404/429/5xx; timeout на странице; unsupported capability; compare response с неполнотой |
| Domain / SQL | Namespace transaction rollback; разные namespace policies; canonical package names; prerelease/range/alias; миграция и восстановление БД |
| Performance | Render фиксированного viewport 100/1000/5000; allocations; cache-read count при N refresh; API request count на project snapshot |

Полезные существующие основы: generic selection helpers в `internal/app/select.go`, общий registry `ProjectChecks/ProjectOperations`, транзакция `ReplaceForProjectRun`, компоненты оформления и текущие Go-тесты. Их следует развивать, сохраняя проверенное поведение.

## 9. Дополнительные вопросы и границы вывода

- Проверить контракт SQLCipher-драйвера отдельно: приложение передаёт `_pragma_kdf_iter`, но передача параметра сама по себе не доказывает его применение. Обработку passphrase с кавычками, effective KDF, existing-directory permissions, ошибку `Chmod` и Windows ACL проверить на тестовой БД; текущая ревизия игнорирует ошибку `Chmod` (`storage/store.go:50–72,136–145`). Это не объявлено доказанным обходом шифрования.
- Разделить Source для UI и credential material: сейчас `sources` содержит PAT и проходит через общую модель/renderer. Фактической утечки через экран не найдено — поле маскируется, — но представлению достаточно secret reference/наличия токена.
- Документировать отправку dependency metadata внешним сервисам и поддерживаемые версии npm/OSV/Trivy. Передача собственного VCS PAT в OSV constructors не обнаружена; inherited environment внешних scanner tools требует отдельного решения.
- Добавить README с quick start, требованиями к terminal/font, матрицей provider/stack/scanner, назначением `r/R`, credential scopes и типовыми ошибками. `make dev` сейчас привязан к macOS/iTerm; это удобство разработчика, а не кроссплатформенная команда запуска (`Makefile:22–37`).
- Зафиксировать автоматические проверки в принятой системе CI: тесты, vet, build, позже parser fuzzing и сценарии UI. В текущем tracked inventory конфигурации CI не найдено; внешняя настройка CI не проверялась.
- Не выдавать отсутствие результатов сканирования за доказательство безопасности проекта. Не считать сам факт старой версии библиотеки подтверждённой уязвимостью без применимого advisory и анализа использования.

## 10. Проверочные артефакты

Дополнительный security report, исходные обзоры, 11 reproductions и benchmark сохранены в каталоге текущего Codex Security scan:

```text
/Users/kolosov.a/.codex/state/plugins/codex-security/scans/tld/
bd1ef70996abd20382757c02a84e3f5143609473_20261001T111438Z_276ntys4/
```

Ключевые файлы внутри: `report.md`, `artifacts/01_context/architecture-review.json`, `artifacts/02_discovery/baseline-review.json`, `artifacts/03_validation/audit-repro.log`, `artifacts/03_validation/audit-benchmark.log` и четыре `*_audit_repro_test.go`. Security scan завершён: четыре подтверждённых security findings (1 Medium, 3 Low); полнота обзора тестовых исходников отмечена как partial. Это локальные артефакты проверки; для чтения данного плана они не обязательны. Пути и номера строк в документе относятся к указанной базовой ревизии. Предупреждение сканера об изменении working tree связано с добавлением этого MD; tracked исходники не изменились.

Служебная метрика, возвращённая Codex Security: 11 261 865 суммарных токенов по трём потокам, из них 11 200 735 input (10 719 744 cached input) и 61 130 output. Это суммарный учёт обращений модели, не размер документа. Advisory-проверка TAC вернула `not_granted` и не блокировала локальную проверку; заявка на расширенный доступ — [Trusted Access for Cyber](https://chatgpt.com/cyber).

Приоритет первого небольшого набора исправлений: **job lifecycle → async identity → source-aware cache → scan coverage → parser/terminal input → selection/submit/error UX**. После закрепления этих инвариантов извлечение общих компонентов будет уменьшать число ошибок, а не только размеры файлов.

## 11. Статус реализации после запроса на выполнение плана

Обновлено **2 октября 2026**. Локальная реализация плана R01–R27 завершена в рабочем дереве ветки `refactor`, без коммита, подключения реальных источников и открытия пользовательской БД. Разделы выше сохранены как исходный аудит и критерии приёмки. Проверка на реальных GitLab/npm/OSV/Trivy и миграция пользовательского файла остаются эксплуатационной приёмкой владельца окружения, а не скрытым локальным шагом.

| Пункты | Реализовано и локально проверено | Осталось до исходной приёмки |
|---|---|---|
| R01–R03 | Sync не зависит от текущего selection; list/detail/latest replies привязаны к stable ID и generation; cache key v2 включает source ID/origin, provider project ID и полный SHA; неоднозначные старые project keys удаляются | Прогнать миграцию на копии пользовательской БД |
| R04–R06 | Clean отличается от no coverage/error; last attempt хранит `complete/unsupported/failed/canceled`, last-good становится `STALE`; OSV требует полного exact package coverage, Trivy считает targets, npm показывает размер dependency graph при наличии metadata; revision/time/provenance видны; scanner output ограничен 64 MiB; Gradle EOF и map notation исправлены; C0/C1/ESC удаляются до styling | Зафиксировать версии scanner binaries и большой реальный отчёт в целевом окружении |
| R07–R11 | Stable-ID viewport, PageUp/PageDown/Home/End, правильный CVE selection, Space в input, защита от двойного submit, `/` picker; каждый экран хранит `loading/ready/stale/error` и признак last-good snapshot; ошибка не стирается несвязанным успешным ответом, `Ctrl+R` повторяет чтение | Визуальная приёмка набора терминалов и реальных длинных данных |
| R12–R16 | Root context, deadlines и context-aware каналы; dependencies/checks/releases/vulnerabilities/pins используют один runner, bounded concurrency 4, per-item progress и aggregation; Esc отменяет все пять сценариев, run ID отсекает поздние ответы; checks публикуются одним snapshot, cache reads пакетные; cache TTL и атомарный бюджет 256 MiB; active-screen/visible-row rendering | Выбрать production SLO и снять p95 на целевой машине; текущие microbenchmarks приведены ниже |
| R17–R18 | Pagination реализована для tags и GitLab protection/schedules с page budget и same-origin для внешнего `next`; Namespace + policy сохраняются транзакционно; перед schema migration создаётся backup `0600`, тест v14→v15 восстанавливает копию | Проверить 401/403/429/5xx на реальных провайдерах и ручной restore пользовательского backup |
| R19–R22 | Renderer принимает типизированные screen states; load/action/job contracts вынесены из позиционного root API; сравнение версий едино и строго; package identity использует ecosystem + exact registry name, migration v15 backfill; policy version вычисляется для проекта; missing/error/not-applicable различаются; checks публикуются атомарно | Старые display-name aliases нельзя безопасно угадать: их нужно перечислить на копии реальной БД и исправить явно; provider snapshot не может быть атомарным между несколькими remote endpoints |
| R23–R27 | Screen, navigation, focus и navigation/operations modal actions читаются из registries; hints используют те же bindings; Settings имеет горизонтальное окно, при `<72` двухпанельные экраны показывают активную панель, modal держит selection в вертикальном окне; mutation preview содержит source/project/change, результат типизирован как `applied/partial/uncertain/failed` и подтверждается readback; URL и redirects same-origin; токены отделены от render state; Maintainer показывается по каждому GitLab project; README содержит provider/scanner matrix | Live-проверка прав и mutations допускается только владельцем тестового GitLab; автоматическая компенсация не выполняется, потому что после timeout она сама могла бы повторить уже применённую запись |

Повторяющиеся сценарии U1–U12 закреплены общими primitives: stable selection/viewport, picker, guarded form submit, job lifecycle, load state, action registry, safe text и source-aware snapshot identity. Специализированные таблицы Releases/Vulnerabilities сохраняют собственные renderers, поскольку их timeline/CVE contract отличается от CRUD-списка. Переезд файлов в предложенные каталоги из раздела 6 не выполнялся механически: сам раздел определяет направление зависимостей, а не обязательную смену структуры; нужные границы появились у потребителей (`sourceurl`, `safetext`, version comparison, job/action/load/render contracts).

Для существующей БД есть **совместимость отображения, требующая проверки**: у зависимостей, созданных до явного `registry_name`, это поле могло содержать только display name. После удаления эвристики они будут показаны без фактической версии, если package name отличается (например, `React Native` и `react-native`). До эксплуатации на реальной БД нужно сделать резервную копию, перечислить такие записи и заполнить точное `registry_name` через форму/миграцию; автоматическое угадывание по удалённой пунктуации снова создало бы коллизии. Реальная БД в этой работе не открывалась.

Старые записи checks, сохранённые по одной ячейке, **не выдаются за цельный snapshot**: до первого полного повторного запуска проверок Settings/Dashboard покажут их как неизвестные. Новая запись публикуется одним JSON-элементом кэша только после успешного окончания всех проверок; ошибка посередине сохраняет предыдущий цельный snapshot.

Ограничение кэша считает только `length(value)` и теперь применяется при открытии БД и атомарно при каждой записи; запись больше бюджета отклоняется. Оно не гарантирует размер файла SQLite на диске: удалённые страницы могут остаться в файле до обслуживания БД. Микрозамер `BenchmarkCacheSetWithBudget` на Apple M3 Pro с зашифрованной тестовой БД дал около 63 мкс/запись при 1000 элементах и 316 мкс/запись при 10 000; это конкретная машина и маленький payload, не p95. Стоимость растёт с числом элементов из-за `SUM(length(value))`, поэтому для значительно большего кэша потребуется поддерживаемый счётчик или отдельная политика обслуживания SQLCipher-файла.

Проверки текущего рабочего дерева: `go test ./...`, `go test -race ./...`, `go vet ./...`, `git diff --check` и `go build -o /tmp/tld-refactor-build ./cmd/tld` прошли. Для тестовых HTTP-серверов и повторной чистой сборки потребовалось разрешение вне песочницы. Микрозамер `Projects` на 5000 строках после visible-only rendering: около 0,42 мс и 7,6 тыс. аллокаций на рендер против исходных 57,88 мс и 1,08 млн; это benchmark конкретной сцены, не p95 приложения.
