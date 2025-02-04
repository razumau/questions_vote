import logging
import sqlite3
import os
import datetime

from dotenv import load_dotenv

from db.std_dev import create_stddev_function

logger = logging.getLogger(__name__)

load_dotenv()

DB_PATH = os.getenv("DB_PATH")


def adapt_date_iso(val):
    """Adapt datetime.date to ISO 8601 date."""
    return val.isoformat()


def adapt_datetime_iso(val):
    """Adapt datetime.datetime to timezone-naive ISO 8601 date."""
    return val.isoformat()


def adapt_datetime_epoch(val):
    """Adapt datetime.datetime to Unix timestamp."""
    return int(val.timestamp())


sqlite3.register_adapter(datetime.date, adapt_date_iso)
sqlite3.register_adapter(datetime.datetime, adapt_datetime_iso)
sqlite3.register_adapter(datetime.datetime, adapt_datetime_epoch)


def convert_date(val):
    """Convert ISO 8601 date to datetime.date object."""
    return datetime.date.fromisoformat(val.decode())


def convert_datetime(val):
    """Convert ISO 8601 datetime to datetime.datetime object."""
    return datetime.datetime.fromisoformat(val.decode())


def convert_timestamp(val):
    """Convert Unix epoch timestamp to datetime.datetime object."""
    return datetime.datetime.fromtimestamp(int(val))


sqlite3.register_converter("date", convert_date)
sqlite3.register_converter("datetime", convert_datetime)
sqlite3.register_converter("timestamp", convert_timestamp)


_connection = None


def connection():
    global _connection
    if _connection is None:
        _connection = sqlite3.connect(DB_PATH, detect_types=sqlite3.PARSE_DECLTYPES)
    return _connection


def setup_database():
    logger.info("Creating tables if necessary...")
    with connection() as conn:
        cursor = conn.cursor()

        cursor.execute("""
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
        """)

        # Create images table with foreign key to questions
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS images (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                question_id INTEGER,
                image_url TEXT NOT NULL,
                data BLOB,
                mime_type TEXT,
                FOREIGN KEY (question_id) REFERENCES questions (id)
            )
        """)

        cursor.execute("""
            CREATE TABLE IF NOT EXISTS votes (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                tournament_id INTEGER NOT NULL,
                question1_id INTEGER NOT NULL,
                question2_id INTEGER NOT NULL,
                selected_id INTEGER,
                timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (question1_id) REFERENCES questions (id),
                FOREIGN KEY (question2_id) REFERENCES questions (id)
            )
        """)

        cursor.execute("""
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
        """)

        cursor.execute("""
            CREATE TABLE IF NOT EXISTS tournaments (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    title TEXT NOT NULL,
                    state INTEGER DEFAULT 0,
                    initial_k REAL DEFAULT 64.0,
                    minimum_k REAL DEFAULT 16.0,
                    std_dev_multiplier REAL DEFAULT 2.0,
                    initial_phase_matches INTEGER DEFAULT 10,
                    transition_phase_matches INTEGER DEFAULT 20,
                    top_n INTEGER DEFAULT 100,
                    questions_count INTEGER NOT NULL,
                    band_size INTEGER DEFAULT 200,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
                )
                """)
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS tournament_questions (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                tournament_id INTEGER NOT NULL,
                question_id INTEGER NOT NULL,
                rating REAL NOT NULL,
                matches INTEGER DEFAULT 0,
                wins INTEGER DEFAULT 0
            )
        """)

        create_stddev_function(conn)
        conn.commit()
    logger.info("Tables created")


def clean_database():
    with sqlite3.connect(DB_PATH) as conn:
        cursor = conn.cursor()
        cursor.execute("""DELETE FROM packages""")
        cursor.execute("""DELETE FROM images""")
        cursor.execute("""DELETE FROM questions""")
        cursor.execute("""DELETE FROM votes""")
        conn.commit()


def drop_tables():
    with sqlite3.connect(DB_PATH) as conn:
        cursor = conn.cursor()
        cursor.execute("""DROP TABLE IF EXISTS packages""")
        cursor.execute("""DROP TABLE IF EXISTS images""")
        cursor.execute("""DROP TABLE IF EXISTS questions""")
        cursor.execute("""DROP TABLE IF EXISTS votes""")
        conn.commit()


if __name__ == "__main__":
    setup_database()
    # drop_tables()
    # clean_database()
