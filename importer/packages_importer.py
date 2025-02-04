import os

from db import connection
from models import Question
from package_parser import PackageParser
from utils import sleep_around

DB_PATH = os.getenv("DB_PATH")


def filter_by_year(year):
    with connection() as conn:
        cursor = conn.cursor()
        cursor.execute("SELECT gotquestions_id, end_date FROM packages")
        rows = cursor.fetchall()
        return [row for row in rows if row[1].year == year]


packages_for_year = filter_by_year(2020)
for index, row in enumerate(packages_for_year):
    package_id = row[0]
    if Question.has_questions_from_package(package_id):
        print(f"Skipping package #{index}/{len(packages_for_year)}, id {package_id} (already has questions)")
        continue
    print(f"Processing package #{index}/{len(packages_for_year)}, id {package_id}")
    PackageParser(package_id).import_package()
    sleep_around(1, 0.7)
