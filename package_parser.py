from itertools import chain

import requests

from db import connection, setup_database
from models import Question
from nextjs_helper import extract_next_props

class PackageParser:
    def __init__(self, package_id: int, rewrite: bool = True):
        self.package_id = package_id
        self.url = f'https://gotquestions.online/pack/{package_id}'
        self.rewrite = rewrite

    def import_package(self):
        self.maybe_drop_old_entries()

        response = requests.get(self.url)
        if response.status_code != 200:
            print(f"Error: Unable to fetch URL. Status code: {response.status_code}")
            return

        props = extract_next_props(response)
        if not props:
            return
        try:
            tours = props['props']['pageProps']['pack']['tours']
            all_questions = chain.from_iterable(tour['questions'] for tour in tours)
        except KeyError:
            all_questions = props['props']['pageProps']['tour']['questions']
        for question_dict in all_questions:
            question = Question.build_question(question_dict)
            self.insert_question(question)

    def maybe_drop_old_entries(self):
        if self.rewrite:
            Question.delete_all_questions_for_package(self.package_id)

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
                image_url = f'https://gotquestions.online/{question.handout_img}'
                response = requests.get(image_url)
                image_data = response.content
                mime_type = response.headers.get('Content-Type')
                cursor.execute('''
                    INSERT INTO images (question_id, image_url, data, mime_type) VALUES (?, ?, ?, ?)
                ''', (question_id, question.handout_img, image_data, mime_type))

            conn.commit()


if __name__ == '__main__':
    setup_database()
    parser = PackageParser(5220, rewrite=True)
    parser.import_package()
