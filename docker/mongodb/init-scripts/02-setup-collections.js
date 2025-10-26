db.getSiblingDB('admin').auth('admin', 'mongosecret');

db = db.getSiblingDB('cryptodb');

// Создаем коллекции (автоматически создаются при первой записи)
db.createCollection('contacts');

