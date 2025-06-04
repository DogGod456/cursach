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

-- Таблица участников чата (связь между чатами и пользователями)
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
    id_user UUID NOT NULL,  -- отправитель
    message_text TEXT NOT NULL,
    id_reply_message UUID,  -- ответ на сообщение
    draft BOOLEAN NOT NULL DEFAULT FALSE,
    sending_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- Таблица для отозванных токенов
CREATE TABLE IF NOT EXISTS revoked_tokens (
    id_revoked_token UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token TEXT NOT NULL UNIQUE,
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы
CREATE INDEX idx_chat_users_chat ON chat_users(id_chat);
CREATE INDEX idx_chat_users_user ON chat_users(id_user);
CREATE INDEX idx_messages_chat ON messages(id_chat);
CREATE INDEX idx_messages_sender ON messages(id_user);
CREATE INDEX idx_messages_time ON messages(sending_time);
CREATE INDEX idx_revoked_tokens_token ON revoked_tokens(token);

-- Внешние ключи (все связи описаны после создания таблиц)
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

ALTER TABLE messages 
ADD CONSTRAINT fk_messages_reply 
FOREIGN KEY (id_reply_message) REFERENCES messages(id_message) ON DELETE SET NULL;

`
