from dataclasses import dataclass

import apsw.ext

from db import connection
from models.tournament_question import TournamentQuestion

INITIAL_RATING = 1500.0


@dataclass
class Tournament:
    id: int
    questions_count: int
    title: str
    initial_k: float
    minimum_k: float
    std_dev_multiplier: float
    initial_phase_matches: int
    transition_phase_matches: int
    top_n: int
    band_size: int

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
        initial_rating: float = INITIAL_RATING,
        band_size: int = 200,
    ):
        tournament_id = (
            connection()
            .execute(
                """
            INSERT INTO tournaments 
            (title, initial_k, minimum_k, std_dev_multiplier, 
                initial_phase_matches, transition_phase_matches, top_n, 
                questions_count, band_size
            )
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            RETURNING id
        """,
                (
                    title,
                    initial_k,
                    minimum_k,
                    std_dev_multiplier,
                    initial_phase_matches,
                    transition_phase_matches,
                    top_n,
                    len(question_ids),
                    band_size,
                ),
            )
            .fetchone()[0]
        )

        TournamentQuestion.create_tournament_questions(tournament_id, question_ids, initial_rating)

    @classmethod
    def start_tournament(cls, title):
        connection().execute("UPDATE tournaments SET state = 1 WHERE title = ?", (title,))

    @classmethod
    def list_active_tournaments(cls):
        conn = connection()
        conn.row_trace = apsw.ext.DataClassRowFactory(dataclass_kwargs={"frozen": True})

        rows = conn.execute("SELECT * FROM tournaments WHERE state = 1").fetchall()
        return [
            cls(
                id=row.id,
                title=row.title,
                initial_k=row.initial_k,
                minimum_k=row.minimum_k,
                std_dev_multiplier=row.std_dev_multiplier,
                initial_phase_matches=row.initial_phase_matches,
                transition_phase_matches=row.transition_phase_matches,
                top_n=row.top_n,
                questions_count=row.questions_count,
                band_size=row.band_size,
            )
            for row in rows
        ]

    @classmethod
    def find_active_tournament(cls):
        conn = connection()
        conn.row_trace = apsw.ext.DataClassRowFactory(dataclass_kwargs={"frozen": True})
        rows = conn.execute("SELECT * FROM tournaments WHERE state = 1").fetchall()
        if len(rows) != 1:
            raise ValueError(f"There should be exactly one active tournament, not {len(rows)}")

        row = rows[0]
        return cls(
            id=row.id,
            title=row.title,
            initial_k=row.initial_k,
            minimum_k=row.minimum_k,
            std_dev_multiplier=row.std_dev_multiplier,
            initial_phase_matches=row.initial_phase_matches,
            transition_phase_matches=row.transition_phase_matches,
            top_n=row.top_n,
            questions_count=row.questions_count,
            band_size=row.band_size,
        )

    def reset_tournament(self):
        connection().execute(
            """update tournament_questions set rating = ?, wins = 0, matches = 0 
               where tournament_id = ?
           """,
            (INITIAL_RATING, self.id),
        )
