from dataclasses import dataclass
from itertools import chain

import requests

from db import connection, setup_database, drop_tables
from nextjs_helper import extract_next_props

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
    author_id: int
    package_id: int
    difficulty: float
    is_incorrect: bool

    @classmethod
    def build_question(cls, question_dict):
        return cls(got_questions_id=question_dict['id'],
                   question=question_dict['text'],
                   answer=question_dict['answer'],
                   accepted_answer=question_dict['zachet'],
                   comment=question_dict['comment'],
                   handout_str=question_dict['razdatkaText'],
                   handout_img=question_dict['razdatkaPic'],
                   source=question_dict['source'],
                   author_id=question_dict.get('authors', [])[0]['id'],
                   package_id=question_dict.get('packId', -1),
                   difficulty=question_dict['complexity'],
                   is_incorrect=question_dict['takenDown']
                   )


class PackageParser:
    def __init__(self, package_id: int):
        self.url = f'https://gotquestions.online/pack/{package_id}'

    def import_package(self):
        response = requests.get(self.url)
        if response.status_code != 200:
            print(f"Error: Unable to fetch URL. Status code: {response.status_code}")
            return

        props = extract_next_props(response)
        if not props:
            return
        tours = props['props']['pageProps']['pack']['tours']
        all_questions = chain.from_iterable(tour['questions'] for tour in tours)
        for question_dict in all_questions:
            question = Question.build_question(question_dict)
            self.insert_question(question)

    @staticmethod
    def insert_question(question: Question):
        with connection() as conn:
            cursor = conn.cursor()

            cursor.execute('''
                INSERT INTO questions (
                    gotquestions_id, question, answer, accepted_answer, comment, handout_str, source, author_id, 
                    package_id, difficulty, is_incorrect
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ''', (question.got_questions_id, question.question, question.answer, question.accepted_answer,
                  question.comment, question.handout_str, question.source, question.author_id, question.package_id,
                  question.difficulty, question.is_incorrect))

            if question.handout_img:
                question_id = cursor.lastrowid
                cursor.execute('''INSERT INTO images (question_id, image_url) VALUES (?, ?)''',
                               (question_id, question.handout_img))

            conn.commit()


if __name__ == '__main__':
    drop_tables()
    setup_database()
    parser = PackageParser(6124)
    parser.import_package()
