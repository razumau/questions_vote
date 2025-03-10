from dataclasses import dataclass
from datetime import datetime
from typing import Optional

import apsw.ext

from db import connection


@dataclass
class Package:
    gotquestions_id: int
    title: str
    start_date: datetime
    end_date: datetime
    questions_count: int
    id: Optional[int] = None

    @classmethod
    def build_package(cls, package_dict):
        return cls(
            gotquestions_id=package_dict["id"],
            title=package_dict["title"],
            start_date=datetime.fromisoformat(package_dict["startDate"]),
            end_date=datetime.fromisoformat(package_dict["endDate"]),
            questions_count=package_dict["questions"],
        )

    @classmethod
    def find_package(cls, gotquestions_id):
        conn = connection()
        conn.row_trace = apsw.ext.DataClassRowFactory(dataclass_kwargs={"frozen": True})
        row = conn.execute("SELECT * FROM packages WHERE gotquestions_id = ?", (gotquestions_id,)).fetchone()
        if row:
            return cls(
                id=row.id,
                gotquestions_id=row.gotquestions_id,
                title=row.title,
                start_date=row.start_date,
                end_date=row.end_date,
                questions_count=row.questions_count,
            )
        return None
