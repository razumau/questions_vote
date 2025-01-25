import sqlite3
from dataclasses import dataclass

from db import connection


@dataclass
class Tournament:
    id: int
    initial_k: float = 64.0
    minimum_k: float = 16.0
    std_dev_multiplier: float = 2.0
    initial_phase_matches: int = 10
    transition_phase_matches: int = 20
    top_n: int = 100

    @classmethod
    def create_tournament(
        cls,
        question_ids: list[int],
        title: str,
        initial_k: float = 64.0,
        minimum_k: float = 16.0,
        std_dev_multiplier: float = 2.0,
        initial_phase_matches: int = 10,
        transition_phase_matches: int = 20,
        top_n: int = 100,
        initial_rating: float = 1500.0,
    ):
        with connection() as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                INSERT INTO tournaments 
                (title, initial_k, minimum_k, std_dev_multiplier, initial_phase_matches, transition_phase_matches, top_n)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            """,
                (title, initial_k, minimum_k, std_dev_multiplier, initial_phase_matches, transition_phase_matches, top_n),
            )

            conn.commit()
            tournament_id = cursor.lastrowid

            cursor.executemany(
                """
                INSERT INTO tournament_questions (tournament_id, question_id, rating) VALUES (?, ?, ?)
                """,
                [(tournament_id, question_id, initial_rating) for question_id in question_ids],
            )
            conn.commit()

    @classmethod
    def find_active_tournament(cls):
        with connection() as conn:
            conn.row_factory = sqlite3.Row
            cursor = conn.cursor()
            cursor.execute("SELECT * FROM tournaments WHERE state = 1")
            rows = cursor.fetchall()
            if len(rows) != 1:
                return None

            row = rows[0]
            return cls(
                id=row["id"],
                initial_k=row["initial_k"],
                minimum_k=row["minimum_k"],
                std_dev_multiplier=row["std_dev_multiplier"],
                initial_phase_matches=row["initial_phase_matches"],
                transition_phase_matches=row["transition_phase_matches"],
                top_n=row["top_n"],
            )

