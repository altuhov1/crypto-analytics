db.getSiblingDB('admin').auth('admin', 'mongosecret');

db = db.getSiblingDB('cryptodb');

db.createCollection('posts');
db.createCollection('comments');

