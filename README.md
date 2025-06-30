## Структура проекта
- **wbtech** — основной сервис
- **emulator** — вспомогательный сервис (отвечает за отправку сообщений)
- **l0** — UI

## Запуск проекта
Для запуска приложения выполните команды:
```sh
git clone git@github.com:IlyaAGL/wb-l0.git
cd wb-l0
docker compose -f wbtech/docker-compose.yml up -d
docker compose -f emulator/docker-compose.yml up -d
```
