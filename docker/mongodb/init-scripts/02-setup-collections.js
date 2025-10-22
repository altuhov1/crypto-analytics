db.getSiblingDB('admin').auth('admin', 'mongosecret');

db = db.getSiblingDB('cryptodb');

// Создаем коллекции (автоматически создаются при первой записи)
db.createCollection('contacts');
db.createCollection('users');
db.createCollection('transactions');

// Создаем индексы для оптимизации
db.contacts.createIndex({ "email": 1 }, { unique: true });
db.contacts.createIndex({ "created_at": -1 });

db.users.createIndex({ "username": 1 }, { unique: true });
db.users.createIndex({ "email": 1 }, { unique: true });

db.transactions.createIndex({ "user_id": 1, "created_at": -1 });
db.transactions.createIndex({ "type": 1 });

// Вставляем тестовые данные (опционально)
db.contacts.insertOne({
    name: "Test User",
    email: "test@example.com",
    message: "Initial test message",
    created_at: new Date()
});

print("MongoDB initialization completed successfully!");