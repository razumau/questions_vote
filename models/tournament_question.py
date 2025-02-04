from dataclasses import dataclass
from random import randint
from typing import Optional

from db import connection


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
    def get_random_question(
        cls,
        tournament_id: int,
        min_rating: float = 0.0,
        max_rating: float = 1_000_000.0,
        max_matches: int = 1_000_000,
    ) -> Optional["TournamentQuestion"]:
        with connection() as conn:
            cursor = conn.cursor()
            questions_count = cursor.execute(
                """
                select count(*) 
                from tournament_questions 
                where tournament_id = ? 
                    and rating between ? and ?
                    and matches <= ?
                """,
                (tournament_id, min_rating, max_rating, max_matches),
            ).fetchone()[0]

            offset = randint(0, questions_count - 1)
            row = cursor.execute(
                """
                select question_id, rating
                from tournament_questions
                where tournament_id = ? 
                    and rating between ? and ?
                    and matches <= ?
                limit 1 
                offset ? 
                """,
                (tournament_id, min_rating, max_rating, max_matches, offset),
            ).fetchone()

            return cls(tournament_id=tournament_id, question_id=row[0], rating=row[1])

    @classmethod
    def get_qualified_ratings(cls, tournament_id: int, initial_phase_matches: int) -> list[int]:
        with connection() as conn:
            cursor = conn.cursor()
            rows = cursor.execute(
                """
                select rating
                from tournament_questions
                where tournament_id = ? and matches >= ?
                """,
                (tournament_id, initial_phase_matches),
            ).fetchall()

        return [row[0] for row in rows]

    @classmethod
    def get_rating_at_position(cls, tournament_id: int, position: int) -> int:
        with connection() as conn:
            cursor = conn.cursor()
            rows = cursor.execute(
                """
                select rating
                from tournament_questions
                where tournament_id = ?
                order by rating desc
                limit 1 offset ?
                """,
                (tournament_id, position),
            ).fetchone()

        return rows[0]

    @classmethod
    def get_stats_for_qualified(cls, tournament_id: int, initial_phase_matches: int) -> tuple[int, float]:
        with connection() as conn:
            cursor = conn.cursor()
            row = cursor.execute(
                """
                select count(rating), stddev(rating)
                from tournament_questions
                where tournament_id = ? and matches >= ?
                """,
                (tournament_id, initial_phase_matches),
            ).fetchone()

        return row

    @classmethod
    def get_rating_distribution(cls, tournament_id: int, bin_size: int = 20) -> dict:
        with connection() as conn:
            cursor = conn.cursor()
            rows = cursor.execute(
                """
                select round(rating / ?) * ? as bin, count(*) as count
                from tournament_questions
                where tournament_id = ?
                group by bin
                order by bin
                """,
                (bin_size, bin_size, tournament_id),
            ).fetchall()

            return {int(row[0]): row[1] for row in rows}

    @classmethod
    def count_questions_above_threshold(cls, tournament_id: int, threshold: float) -> int:
        with connection() as conn:
            cursor = conn.cursor()
            row = cursor.execute(
                """
                select count(*)
                from tournament_questions
                where tournament_id = ? and rating >= ?
                """,
                (tournament_id, threshold),
            ).fetchone()

            return row[0]

    @classmethod
    def count_unqualified_questions(cls, tournament_id: int, qualification_cutoff: int) -> int:
        with connection() as conn:
            cursor = conn.cursor()
            row = cursor.execute(
                """
                select count(*)
                from tournament_questions
                where tournament_id = ? and matches < ?
                """,
                (tournament_id, qualification_cutoff),
            ).fetchone()

            return row[0]

    @classmethod
    def get_top_questions(cls, tournament_id: int, n: int) -> list["TournamentQuestion"]:
        with connection() as conn:
            cursor = conn.cursor()
            rows = cursor.execute(
                """
                select question_id, rating, matches, wins
                from tournament_questions
                where tournament_id = ?
                order by rating desc
                limit ?
                """,
                (tournament_id, n),
            ).fetchall()

            return [
                cls(tournament_id=tournament_id, question_id=row[0], rating=row[1], matches=row[2], wins=row[3])
                for row in rows
            ]

    @classmethod
    def get_match_counts(cls, tournament_id: int):
        with connection() as conn:
            cursor = conn.cursor()
            rows = cursor.execute(
                """
                select sum(matches), sum(wins)
                from tournament_questions
                where tournament_id = ?
                """,
                (tournament_id,),
            ).fetchone()

            return {"matches": rows[0], "wins": rows[1]}

    @classmethod
    def find(cls, tournament_id: int, question_id: int) -> Optional["TournamentQuestion"]:
        with connection() as conn:
            cursor = conn.cursor()
            row = cursor.execute(
                """
                select rating, matches, wins
                from tournament_questions
                where tournament_id = ? and question_id = ?
                """,
                (tournament_id, question_id),
            ).fetchone()

            if row is None:
                return None

            return cls(tournament_id=tournament_id, question_id=question_id, rating=row[0], matches=row[1], wins=row[2])

    def save(self):
        with connection() as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                update tournament_questions
                set rating = ?, matches = ?, wins = ?
                where tournament_id = ? and question_id = ?
                """,
                (self.rating, self.matches, self.wins, self.tournament_id, self.question_id),
            )
            conn.commit()
