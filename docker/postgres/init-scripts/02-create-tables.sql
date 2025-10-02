\c webdev;

-- Создаем таблицу (пока без OWNER)
CREATE TABLE contacts (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Меняем владельца таблицы на webuser
ALTER TABLE contacts OWNER TO webuser;

-- Даем права
GRANT ALL PRIVILEGES ON TABLE contacts TO webuser;
GRANT ALL PRIVILEGES ON SEQUENCE contacts_id_seq TO webuser;
GRANT ALL PRIVILEGES ON SCHEMA public TO webuser;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO webuser;