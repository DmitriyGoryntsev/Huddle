# Huddle Web

Фронтенд приложения Huddle — поиск компании для игр и мероприятий.

## Стек

- React 18 + TypeScript
- Vite
- Tailwind CSS
- React Router
- TanStack Query
- Zustand
- Leaflet (карта)

## Разработка

```bash
npm install
npm run dev
```

Приложение: http://localhost:3000

API проксируется на http://localhost (Nginx gateway). Запустите backend: `docker-compose up` в корне проекта.

## Сборка

```bash
npm run build
```

Статика в `dist/`.
