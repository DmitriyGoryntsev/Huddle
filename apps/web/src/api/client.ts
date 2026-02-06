import axios from 'axios';

// Твой Nginx слушает на 80 порту того же хоста
const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Добавляем токен из localStorage, если он есть
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('huddle_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;