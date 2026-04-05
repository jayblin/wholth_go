git submodule update --remote

В .env TZ="" чтобы юзалась UTC-таймзона.

Чтобы запустить нужен файл `.secrets.gpg`, который должен быть получен след.
образом:
1. Скопировать `.secrets.dist` в `.sercrets` и заполнить его;
2. Зашифровать `.secrets.dist`:
```sh
gpg [-u <YOUR_GPG_KEY_ID>] --encrypt .secrets
```
