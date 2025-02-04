from dataclasses import dataclass
from typing import Optional

from db import connection


@dataclass
class Vote:
    @classmethod
    def create(cls, user_id: int, question1_id: int, question2_id: int, tournament_id: int, selected_id: Optional[int]):
        with connection() as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                INSERT INTO votes (user_id, question1_id, question2_id, tournament_id, selected_id)
                VALUES (?, ?, ?, ?, ?)
            """,
                (user_id, question1_id, question2_id, tournament_id, selected_id),
            )
            conn.commit()
