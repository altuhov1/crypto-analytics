db.getSiblingDB('admin').auth('admin', 'mongosecret');

db = db.getSiblingDB('cryptodb');

db.createUser({
    user: 'appuser',
    pwd: 'apppassword',
    roles: [
        {
            role: 'readWrite',
            db: 'cryptodb'
        },
        {
            role: 'dbAdmin',
            db: 'cryptodb'
        }
    ]
});