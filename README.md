﻿# Airflight-service

Это Go HTTP-сервис для бронирования авиабилетов и операций управления, использующий тренировочную базу данных от [PostgresPro](https://postgrespro.ru/docs/postgrespro/15/demodb-bookings). Система реализует REST API, позволяющий клиентам искать рейсы, бронировать маршруты и управлять процессами регистрации с назначением мест.

## Технологический стек

| Компонент |	Технология | Назначение |
| --- | --- | --- |
| HTTP Router |	github.com/go-chi/chi/v5 | Request routing and middleware	|
| Database ORM | gorm.io/gorm	| Object-relational mapping	|
| Database Driver |	gorm.io/driver/postgres | PostgreSQL connectivity |
| Configuration	| github.com/joho/godotenv | Environment variable loading |	
| API Documentation |	github.com/swaggo/swag | OpenAPI specification generation |
| Documentation UI | github.com/swaggo/http-swagger |	Swagger UI serving |

## Назначение сервиса

API Airflight Service служит в качестве внутреннего сервиса для приложений бронирования авиабилетов, обеспечивая функциональность для:
- Поиск информации об аэропортах и городах
- Поиск маршрута полета с возможностью пересадки
- Идемпотентные операции бронирования с использованием транзакций на основе GUID
- Назначение мест и генерация посадочного талона во время регистрации.
