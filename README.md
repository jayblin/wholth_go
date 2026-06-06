HTTP-сервер для движка wholth.
Движок подтягивается как git-submodule.

# Как билдить

## 1. Секреты
Чтобы запустить нужен файл `.secrets.gpg`, который должен быть получен
след. образом:
1. Скопировать `.secrets.dist` в `.sercrets` и заполнить его;
2. Зашифровать `.secrets`:
```sh
gpg [-u <YOUR_GPG_KEY_ID>] --encrypt .secrets
```

## 2. Компиляция
```sh
# компиляция движка wholth и go-сервера;
# подготовка окружения;
# подготовка js-скриптов;
make build
```

## 3. Окружение

Переменные окружения лежат в файле `.env`. Что должны там быть:
```env
# IP-адрес или домен на котором живёт сервер.
DOMAIN=localhost

# Порт по которому доступен сервер.
PORT=8081

# Будут ли go-шаблоны кешироваться: пустое значение - не будут; 1 - будут.
USE_TEMPLATE_CACHE=

# Разрешено ли регистроваться? 1 - да; 0 - нет.
ALLOW_REGISTRATION=1

# Таймзона, в которой работает сервер: пустая строка означает UTC.
TZ=""

# Заполняется автоматичекси при `make build`.
VERSION=36799f2fd3e7000bad19ca1cb7d83bcee29c6c6f
```

## 4. Запуск
```sh
make run
# появится окно для ввода пароля от gpg-ключа. После ввода сервер должен запуститься.
```

# Разработка

## css
Для компиляции css-правил запусти `make css` или `make css-watch`.


# Бэкапы

Бэкапы базы делаю на сервер через такой крон-таск:
```sh
week_ago=$(date --utc --date='now - 1 week' '+%s')

for f_path in $(find <путь до папки с бэкапами> -name "wholth.db*" -type f); do
        birth_timestamp=$(stat --format "%W" $f_path)

        if [ $week_ago -gt $birth_timestamp ]; then
                rm $f_path
        fi
done;

now_ts=$(date --utc '+%s')
sqlite3_rsync <путь до базы> <путь до папки с бэкапами>/wholth.db.$(echo $now_ts).copy
```

`sqlite_rsunc` можно легко скомпилировать:
```sh
git clone https://github.com/sqlite/sqlite.git
cd sqlite
./configure
make sqlite3_rsync
./sqlite3_rsync --help
```
