-- Создаем базу и пользователя (выполняется в postgres базе)
CREATE DATABASE webdev;
CREATE USER webuser WITH PASSWORD '1111';
GRANT ALL PRIVILEGES ON DATABASE webdev TO webuser;