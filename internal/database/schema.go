package database

const Schema = `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id_user UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(10) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- Таблица чатов
CREATE TABLE IF NOT EXISTS chats (
    id_chat UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- Таблица участников чата
CREATE TABLE IF NOT EXISTS chat_users (
    id_chat_user UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL,
    id_user UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS messages (
    id_message UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL,
    id_user UUID NOT NULL,
    message_text TEXT NOT NULL,
    sending_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- Таблица для отозванных токенов
CREATE TABLE IF NOT EXISTS revoked_tokens (
    id_revoked_token UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token TEXT NOT NULL UNIQUE,
    id_user UUID NOT NULL,  
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_chat_users_chat ON chat_users(id_chat);
CREATE INDEX idx_chat_users_user ON chat_users(id_user);
CREATE INDEX idx_messages_chat ON messages(id_chat);
CREATE INDEX idx_messages_sender ON messages(id_user);
CREATE INDEX idx_messages_time ON messages(sending_time);
CREATE INDEX idx_revoked_tokens_token ON revoked_tokens(token);
CREATE INDEX idx_revoked_tokens_user ON revoked_tokens(id_user);  

-- Внешние ключи
ALTER TABLE chat_users 
ADD CONSTRAINT fk_chat_users_chat 
FOREIGN KEY (id_chat) REFERENCES chats(id_chat) ON DELETE CASCADE;

ALTER TABLE chat_users 
ADD CONSTRAINT fk_chat_users_user 
FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE;

ALTER TABLE messages 
ADD CONSTRAINT fk_messages_chat 
FOREIGN KEY (id_chat) REFERENCES chats(id_chat) ON DELETE CASCADE;

ALTER TABLE messages 
ADD CONSTRAINT fk_messages_user 
FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE;

ALTER TABLE revoked_tokens 
ADD CONSTRAINT fk_revoked_tokens_user 
FOREIGN KEY (id_user) REFERENCES users(id_user) ON DELETE CASCADE;

-- Создание роли администратора (полный доступ ко всем таблицам)
CREATE ROLE messenger_admin LOGIN PASSWORD 'Daetoi30';
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO messenger_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO messenger_admin;
GRANT EXECUTE ON FUNCTION uuid_generate_v4() TO messenger_admin;

-- Создание роли обычного пользователя (ограниченный доступ)
CREATE ROLE messenger_user LOGIN PASSWORD 'Daetoi30';
GRANT SELECT, INSERT, UPDATE, DELETE ON 
    users, 
    chats, 
    chat_users, 
    messages 
TO messenger_user;
GRANT EXECUTE ON FUNCTION uuid_generate_v4() TO messenger_user;

-- Явный запрет доступа к таблице revoked_tokens
REVOKE ALL PRIVILEGES ON TABLE revoked_tokens FROM messenger_user;

`
