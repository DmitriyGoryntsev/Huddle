CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER REFERENCES categories(id),
    name VARCHAR(50) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    icon_url TEXT,
    color_code VARCHAR(7)
);

-- 1. Вставляем основные категории (без ручного ID)
INSERT INTO categories (name, slug, color_code) VALUES
('Спорт', 'sports', '#2ecc71'),
('Еда и напитки', 'food_drink', '#e67e22'),
('Развлечения', 'entertainment', '#9b59b6')
ON CONFLICT (slug) DO NOTHING;

-- 2. Вставляем подкатегории, находя parent_id по его slug
INSERT INTO categories (parent_id, name, slug) VALUES
((SELECT id FROM categories WHERE slug = 'sports'), 'Падл-теннис', 'padel'),
((SELECT id FROM categories WHERE slug = 'sports'), 'Футбол', 'football'),
((SELECT id FROM categories WHERE slug = 'food_drink'), 'Бары', 'bars'),
((SELECT id FROM categories WHERE slug = 'entertainment'), 'Бильярд', 'billiards'),
((SELECT id FROM categories WHERE slug = 'entertainment'), 'Настольные игры', 'board-games')
ON CONFLICT (slug) DO NOTHING;