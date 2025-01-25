from dataclasses import dataclass
from random import randint
from typing import Optional

from db import connection

MAX_RATING = 1_000_000.0

@dataclass
class TournamentQuestion:
    tournament_id: int
    question_id: int
    rating: Optional[float] = None
    matches: Optional[int] = None
    wins: Optional[int] = None

    @classmethod
    def create_tournament_questions(cls, tournament_id, question_ids, initial_rating):
        with connection() as conn:
            cursor = conn.cursor()
            cursor.executemany(
                """
                INSERT INTO tournament_questions (tournament_id, question_id, rating) VALUES (?, ?, ?)
                """,
                [(tournament_id, question_id, initial_rating) for question_id in question_ids],
            )
            conn.commit()

    @classmethod
    def get_random_question(cls,
                            tournament_id: int,
                            min_rating: float = 0.0,
                            max_rating: float = MAX_RATING):
        with connection() as conn:
            cursor = conn.cursor()
            questions_count = cursor.execute(
                """
                select count(*) 
                from tournament_questions 
                where tournament_id = ? and rating between ? and ?
                """,
                (tournament_id, min_rating, max_rating),
            ).fetchone()[0]

            offset = randint(0, questions_count - 1)
            row = cursor.execute(
                """
                select question_id, rating
                from tournament_questions
                where tournament_id = ? and rating between ? and ?
                limit 1 
                offset ? 
                """,
                (tournament_id, min_rating, max_rating, offset),
            ).fetchone()

            return cls(tournament_id=tournament_id, question_id=row[0], rating=row[1])

    @classmethod
    def get_qualified_questions(cls, tournament_id: int, initial_phase_matches: int) -> list["TournamentQuestion"]:
        with connection() as conn:
            cursor = conn.cursor()
            rows = cursor.execute(
                """
                select question_id, rating
                from tournament_questions
                where tournament_id = ? and matches >= ?
                """,
                (tournament_id, initial_phase_matches),
            ).fetchall()

            return [cls(tournament_id=tournament_id, question_id=row[0], rating=row[1]) for row in rows]
