-- Добавляем поле display_name в таблицу users
ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name VARCHAR(50);

-- Заполняем существующие записи: display_name = username
UPDATE users SET display_name = username WHERE display_name IS NULL;

-- Для админа можно задать специальное имя
UPDATE users SET display_name = 'Администратор' WHERE username = 'admin' AND display_name = 'admin';

-- Делаем поле NOT NULL после заполнения
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;

-- Можно добавить индекс для поиска
CREATE INDEX IF NOT EXISTS idx_users_display_name ON users(display_name);