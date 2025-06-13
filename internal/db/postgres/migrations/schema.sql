
-- Создание ENUM
CREATE TYPE user_role AS ENUM ('student', 'assistant', 'seminarist', 'lecturer', 'superaccount');

-- 1) USERS
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(32) NOT NULL,
    surname VARCHAR(32) NOT NULL,
    patronymic VARCHAR(32),
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL, -- Хешированный пароль с использованием ARGON2ID
    role user_role DEFAULT 'student' NOT NULL
);


-- 2) student_groups
CREATE TABLE student_groups (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);

-- 3) users_groups
CREATE TABLE users_in_groups (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES student_groups(id) ON DELETE CASCADE
);

-- 4) Disciplines
CREATE TABLE disciplines (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    lector_id BIGINT,
    FOREIGN KEY (lector_id) REFERENCES users(id) ON DELETE SET NULL
);

-- 5) disciplines_in_groups
CREATE TABLE groups_in_disciplines (
    id SERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL,
    discipline_id BIGINT NOT NULL,
	seminarist_id BIGINT NOT NULL, 
	assistant_id BIGINT NOT NULL,
	FOREIGN KEY (seminarist_id) REFERENCES users(id) ON DELETE SET NULL,
	FOREIGN KEY (assistant_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (group_id) REFERENCES student_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (discipline_id) REFERENCES disciplines(id) ON DELETE CASCADE
);



-- Таблицы о работах, созданных лекторами

-- 6) Tasks
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    lector_id BIGINT NOT NULL,
    group_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    deadline TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    discipline_id BIGINT NOT NULL,
    content_url TEXT, -- Ссылка на S3 для материалов задачи
    FOREIGN KEY (lector_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES student_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (discipline_id) REFERENCES disciplines(id) ON DELETE CASCADE
);

-- 7) criteria_groups
CREATE TABLE criteria_groups (
    id SERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    group_name TEXT NOT NULL,
    block_flag BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- 8) criteria
CREATE TABLE criteria (
    id SERIAL PRIMARY KEY,
	name TEXT, 
	description TEXT,
    criteria_group_id BIGINT NOT NULL,
    weight BIGINT NOT NULL CHECK (weight >= 0),
    comment_000 TEXT,
    comment_025 TEXT,
    comment_050 TEXT,
    comment_075 TEXT,
    comment_100 TEXT,
    FOREIGN KEY (criteria_group_id) REFERENCES criteria_groups(id) ON DELETE CASCADE
);

-- Таблица работ студентов

-- 10) works
CREATE TABLE student_works (
    id SERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL CHECK (status IN ('pending', 'submitted', 'graded by assistant', 'graded by seminarist')),
    task_id BIGINT NOT NULL,
	assistant_id BIGINT,
	FOREIGN KEY (assistant_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
   
);

-- 11) student_criteria_marks
CREATE TABLE student_criteria_marks (
    id SERIAL PRIMARY KEY,
    student_work_id BIGINT NOT NULL,
    criteria_id BIGINT NOT NULL,
    mark NUMERIC(3,2) NOT NULL CHECK (mark >= 0 AND mark <= 1), -- Специальный тип для точных значений (0.00 - 1.00)
    comment TEXT,
    FOREIGN KEY (student_work_id) REFERENCES student_works(id) ON DELETE CASCADE,
    FOREIGN KEY (criteria_id) REFERENCES criteria(id) ON DELETE CASCADE
);

-- Остальные таблицы

-- 12) Notifications
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    message TEXT NOT NULL,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);