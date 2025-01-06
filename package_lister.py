import random
import time
from dataclasses import dataclass
from datetime import datetime

import requests

from nextjs_helper import extract_next_props
from db import connection, setup_database, drop_tables, clean_database


@dataclass
class Package:
    gotquestions_id: int
    title: str
    start_date: datetime
    end_date: datetime
    questions_count: int

    @classmethod
    def build_package(cls, package_dict):
        return cls(
            gotquestions_id=package_dict['id'],
            title=package_dict['title'],
            start_date=datetime.fromisoformat(package_dict['startDate']),
            end_date=datetime.fromisoformat(package_dict['endDate']),
            questions_count=package_dict['questions']
        )


class PackageLister:
    def __init__(self, first_page: int = 1, last_page: int = 337):
        self.first_page = first_page
        self.last_page = last_page

    def run(self):
        for page in range(self.first_page, self.last_page + 1):
            if page % 10 == 1:
                print(f"Processing page {page}")
            time.sleep(random.uniform(0.3, 0.7))
            self.create_packages_from_page(page)

    def create_packages_from_page(self, page: int):
        url = f'https://gotquestions.online/packs?page={page}'
        response = requests.get(url)
        if response.status_code != 200:
            print(f"Error: Unable to fetch URL. Status code: {response.status_code}")
            return
        props = extract_next_props(response)
        if not props:
            return
        packages = props['props']['pageProps']['packs']
        for package_dict in packages:
            package = Package.build_package(package_dict)
            self.insert_package(package)

    def insert_package(self, package: Package):
        with connection() as conn:
            cursor = conn.cursor()
            if self.package_exists(package.gotquestions_id):
                return
            cursor.execute('''
                INSERT INTO packages (
                    gotquestions_id, title, start_date, end_date, questions_count
                ) VALUES (?, ?, ?, ?, ?)
            ''', (
                package.gotquestions_id,
                package.title,
                package.start_date,
                package.end_date,
                package.questions_count
            ))
            conn.commit()

    @staticmethod
    def package_exists(package_id) -> bool:
        with connection() as conn:
            cursor = conn.cursor()
            cursor.execute('SELECT * FROM packages WHERE gotquestions_id = ?', (package_id,))
            row = cursor.fetchone()
            return row is not None


if __name__ == '__main__':
    setup_database()
    clean_database()
    parser = PackageLister(1, 100)
    parser.run()
