-- 0001 — начальная схема. Эквивалент того, что генерирует GORM AutoMigrate
-- из текущих моделей. Файл хранится отдельно, чтобы изменения схемы можно
-- было ревьюить как обычный код, а не угадывать по diff'у моделей.

CREATE TABLE IF NOT EXISTS users (
    id                  UUID PRIMARY KEY,
    name                VARCHAR(255),
    email               VARCHAR(255) NOT NULL UNIQUE,
    password_hash       VARCHAR(255) NOT NULL,
    role                VARCHAR(50) DEFAULT 'teacher',
    reset_token         VARCHAR(255),
    reset_token_expiry  TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS quizzes (
    id           UUID PRIMARY KEY,
    teacher_id   UUID NOT NULL,
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    status       VARCHAR(50) DEFAULT 'draft',
    mode         VARCHAR(50) DEFAULT 'teacher_paced',
    settings     JSONB,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_quizzes_teacher_id ON quizzes (teacher_id);

CREATE TABLE IF NOT EXISTS questions (
    id                    UUID PRIMARY KEY,
    quiz_id               UUID NOT NULL REFERENCES quizzes(id) ON UPDATE CASCADE ON DELETE CASCADE,
    type                  VARCHAR(50) NOT NULL,
    text                  TEXT NOT NULL,
    image_url             TEXT,
    timer_sec             INT NOT NULL,
    score                 INT NOT NULL,
    "order"               INT NOT NULL,
    correct_text_answers  JSONB
);
CREATE INDEX IF NOT EXISTS idx_questions_quiz_id ON questions (quiz_id);

CREATE TABLE IF NOT EXISTS options (
    id          UUID PRIMARY KEY,
    question_id UUID NOT NULL REFERENCES questions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    text        TEXT NOT NULL,
    is_correct  BOOLEAN NOT NULL,
    color       VARCHAR(50)
);
CREATE INDEX IF NOT EXISTS idx_options_question_id ON options (question_id);

CREATE TABLE IF NOT EXISTS game_sessions (
    id                UUID PRIMARY KEY,
    quiz_id           UUID NOT NULL,
    host_id           UUID,
    pin               VARCHAR(10) UNIQUE,
    status            VARCHAR(50) DEFAULT 'waiting',
    mode              VARCHAR(50) DEFAULT 'teacher_paced',
    started_at        TIMESTAMPTZ DEFAULT NOW(),
    finished_at       TIMESTAMPTZ,
    report_snapshot   JSONB
);
CREATE INDEX IF NOT EXISTS idx_game_sessions_quiz_id ON game_sessions (quiz_id);
CREATE INDEX IF NOT EXISTS idx_game_sessions_host_id ON game_sessions (host_id);
