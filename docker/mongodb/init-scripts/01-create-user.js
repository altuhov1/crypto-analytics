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
db.createCollection('posts');
db.createCollection('comments');

