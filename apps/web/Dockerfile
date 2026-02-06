FROM node:20-alpine
WORKDIR /app

# Копируем package.json и package-lock.json (если есть)
COPY package*.json ./

# Устанавливаем зависимости
# Используем --legacy-peer-deps на случай конфликтов версий react
RUN npm install --legacy-peer-deps

# Копируем всё остальное
COPY . .

EXPOSE 5173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]