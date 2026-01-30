CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER REFERENCES categories(id), -- Если NULL, то это главная группа
    name VARCHAR(50) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    icon_url TEXT,
    color_code VARCHAR(7) 
);

-- Наполняем иерархию
-- 1. Сначала главные группы
INSERT INTO categories (id, name, slug, color_code) VALUES 
(1, 'Спорт', 'sports', '#2ecc71'),
(2, 'Еда и напитки', 'food_drink', '#e67e22'),
(3, 'Развлечения', 'entertainment', '#9b59b6');

-- 2. Затем подкатегории
INSERT INTO categories (parent_id, name, slug) VALUES 
(1, 'Падл-теннис', 'padel'),
(1, 'Футбол', 'football'),
(2, 'Бары', 'bars'),
(3, 'Бильярд', 'billiards'),
(3, 'Настольные игры', 'board-games');