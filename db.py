import sqlite3
import os

from dotenv import load_dotenv

load_dotenv()

DB_PATH = os.getenv('DB_PATH')

def connection():
    return sqlite3.connect(DB_PATH)

def setup_database():
    with sqlite3.connect(DB_PATH) as conn:
        cursor = conn.cursor()

        cursor.execute('''
            CREATE TABLE IF NOT EXISTS questions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                gotquestions_id INTEGER NOT NULL,
                package_id INTEGER NOT NULL,
                question TEXT NOT NULL,
                answer TEXT NOT NULL,
                accepted_answer TEXT,
                handout_str TEXT,
                comment TEXT,
                source TEXT,
                author_id INTEGER,
                difficulty REAL,
                is_incorrect BOOLEAN,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        ''')

        # Create images table with foreign key to questions
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS images (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                question_id INTEGER,
                image_url TEXT NOT NULL,
                data BLOB,
                mime_type TEXT,
                FOREIGN KEY (question_id) REFERENCES questions (id)
            )
        ''')

        cursor.execute('''
            CREATE TABLE IF NOT EXISTS votes (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                question1_id INTEGER NOT NULL,
                question2_id INTEGER NOT NULL,
                selected_id INTEGER,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (question1_id) REFERENCES questions (id),
                FOREIGN KEY (question2_id) REFERENCES questions (id)
            )
        ''')

        cursor.execute('''
            CREATE TABLE IF NOT EXISTS packages (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    gotquestions_id INTEGER NOT NULL,
                    title TEXT NOT NULL,
                    start_date TIMESTAMP,
                    end_date TIMESTAMP,
                    questions_count INTEGER,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                )
        ''')

        conn.commit()
    print("Tables created")

def clean_database():
    with sqlite3.connect(DB_PATH) as conn:
        cursor = conn.cursor()
        cursor.execute('''DELETE FROM packages''')
        cursor.execute('''DELETE FROM images''')
        cursor.execute('''DELETE FROM questions''')
        cursor.execute('''DELETE FROM votes''')
        conn.commit()

def drop_tables():
    with sqlite3.connect(DB_PATH) as conn:
        cursor = conn.cursor()
        cursor.execute('''DROP TABLE IF EXISTS packages''')
        cursor.execute('''DROP TABLE IF EXISTS images''')
        cursor.execute('''DROP TABLE IF EXISTS questions''')
        cursor.execute('''DROP TABLE IF EXISTS votes''')
        conn.commit()
