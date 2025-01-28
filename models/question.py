import sqlite3
from dataclasses import dataclass
from typing import Optional
from datetime import datetime as dt

from db import connection


@dataclass
class Question:
    got_questions_id: int
    question: str
    answer: str
    accepted_answer: str
    comment: str
    source: str
    handout_str: str
    handout_img: str
    author_id: Optional[int]
    package_id: int
    difficulty: float
    is_incorrect: bool

    @classmethod
    def build_question(cls, question_dict):
        author_id = (question_dict["authors"] or [{"id": None}])[0]["id"]

        return cls(
            got_questions_id=question_dict["id"],
            question=question_dict["text"],
            answer=question_dict["answer"],
            accepted_answer=question_dict["zachet"],
            comment=question_dict["comment"],
            handout_str=question_dict["razdatkaText"],
            handout_img=question_dict["razdatkaPic"],
            source=question_dict["source"],
            author_id=author_id,
            package_id=question_dict.get("packId", -1),
            difficulty=question_dict["complexity"],
            is_incorrect=question_dict["takenDown"],
        )

    @classmethod
    def has_questions_from_package(cls, package_id):
        with connection() as conn:
            cursor = conn.cursor()
            cursor.execute("SELECT COUNT(*) FROM questions WHERE package_id = ?", (package_id,))
            return cursor.fetchone()[0] > 0

    @classmethod
    def delete_all_questions_for_package(cls, package_id):
        with connection() as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                DELETE FROM images 
                WHERE question_id IN (SELECT id FROM questions WHERE package_id = ?)
            """,
                (package_id,),
            )
            cursor.execute("DELETE FROM questions WHERE package_id = ?", (package_id,))
            conn.commit()

    @classmethod
    def question_ids_for_year(cls, year):
        start_timestamp = dt.timestamp(dt(year, 1, 1))
        end_timestamp = dt.timestamp(dt(year + 1, 1, 1))
        with connection() as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                SELECT id FROM questions 
                WHERE package_id IN (
                    SELECT gotquestions_id FROM packages WHERE end_date between ? and ?
                )
                AND is_incorrect = 0
            """,
                (start_timestamp, end_timestamp),
            )
            return [row[0] for row in cursor.fetchall()]

    @classmethod
    def find(cls, ids: list[int]) -> list:
        with connection() as conn:
            cursor = conn.cursor()
            cursor.row_factory = sqlite3.Row
            cursor.execute(
                f"""
                select q.id, q.question, q.handout_str, i.mime_type, i.data as image_data 
                from questions q
                left join images i on q.id = i.question_id
                where q.id in ({','.join('?' * len(ids))})
            """,
                ids,
            )
            return cursor.fetchall()
