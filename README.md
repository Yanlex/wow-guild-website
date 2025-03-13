# World of Warcraft Guild Website

> [!CAUTION]
> NGINX не стартует без BACKEND

>[!CAUTION]
> Создайте сеть `docker network create wowguild`

## Nginx

### Если нужен https:// Nginx + SSL

в docker-compose.yml указать свой домен и реальную почту  
`- --certificatesresolvers.leresolver.acme.email=<< ВАШ ЕМЕЙЛ >>`  
`- "traefik.http.routers.nginx.rule=Host(`<<< ВАШЕ ДОМЕННОЕ ИМЯ >>>`)"`  

`docker compose up --build -d traefik`  

#### Основные настройки NGINX

Настройки вашего http сервера 
`nginx.conf`

Эта строка отвечает за сервер nginx, если в docker файле не меняли названия ее менять не нужно  
` add_header 'Access-Control-Allow-Origin' 'http://yanlex-wow-guild-front-nginx';`
Это строка проксирует трафик на нашу API, если в docker файле не меняли названия ее менять не нужно  
`proxy_pass http://yanlex-wow-guild-updater:3000;`

В нижнем блоке server, меняем **sanyadev.ru** на ваш домен.

```nginx
server {
        if ($host = sanyadev.ru) {
            return 301 https://$host$request_uri;
        } # managed by Certbot

        listen 80;
        server_name sanyadev.ru;
        return 404; # managed by Certbot
    }
```

>[!NOTE]
>ЗАПУСК NGINX    
>`docker compose up --build -d frontend`

## Postgres

>Сервис postgres  

Настройки БД находятся в docker-compose.yml  
Нужно настроить 2 поля
>POSTGRES_USER: user-name  
>POSTGRES_PASSWORD: strong-password

Можно так же установить pgAdmin4  
`docker compose up --build -d pgadmin`

>[!NOTE]
>ЗАПУСК POSTGRES  
>`docker compose up --build -d postgres` 

## BACKEND ( API + Updater)

Натросйки приложения находятся в файле backend-docker.yml  
Дефолтные переменные, меняем на свои.
>Сервис backend
```docker
environment:
      DB_NAME: kvd_guild
      GUILD_REGION: eu
      GUILD_REALM: howling-fjord
      GUILD_NAME: "Ключик в дурку"
      DB_USER: user-name
      DB_PASS: strong-password
      DB_NETWORK: wowguild
      DB_ADDRESS: yanlex-wow-guild-postgres
      HOST_DB_PORT: 5432
```
>[!NOTE]
> ЗАПУСК BACKEND ( API + Updater)  
`docker compose up --build -d backend`

## FRONTEND

`npm install`  

Режим разработки  
`npm run dev`

Билд в прод, файлы будут лежать в /public_html  
`npm run build`

---

### Настройка интервала обновления данных

Делает запросы к стороннему АПИ и сверяет есть ли изменения в данных игрока, например рейтинг м+ или количетсво ачивок.  
В файле main.go  

>_, _ = s.Every(1).Day().At("19:12").Do(updatePlayersHandler)

### Логирование в файлы

Это скорее просто для опыта реализовано минимально.
Используем os.UserHomeDir() и основной путь /kvd/logs/ т.е создаем папку kvd в домашнем котологе пользователя куда будут писаться логи

## API

Работает на 3000 порту
Возвращает список игроков гильдии
/api/get-members
Возвращает Rank, Name, Mythic Rating, Guild, Class
/api/guild-data
Шарим папку с аватарками
/api/avatar/
Шарим папку с классами
/api/class/
