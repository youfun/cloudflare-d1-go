-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    age INTEGER,
    email TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS departments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_departments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    department_id INTEGER,
    FOREIGN KEY(user_id) REFERENCES users(id),
    FOREIGN KEY(department_id) REFERENCES departments(id)
);

-- +migrate Down
DROP TABLE user_departments;
DROP TABLE departments;
DROP TABLE users;
