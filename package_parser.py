import json
from dataclasses import dataclass
from itertools import chain

from bs4 import BeautifulSoup

from db import connection, setup_database


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
                   author_id=question_dict.get('authors', [])[0]['id']
                   )


def insert_question(question: Question):
    with connection() as conn:
        cursor = conn.cursor()

        cursor.execute('''
            INSERT INTO questions (
                gotquestions_id, question, answer, accepted_answer, comment, handout_str, source, author_id
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        ''', (question.got_questions_id, question.question, question.answer, question.accepted_answer,
              question.comment, question.handout_str, question.source, question.author_id))

        if question.handout_img:
            question_id = cursor.lastrowid
            cursor.execute('''INSERT INTO images (question_id, image_url) VALUES (?, ?)''',
                           (question_id, question.handout_img))

        conn.commit()


def extract_next_props(html_content):
    soup = BeautifulSoup(html_content, 'html.parser')
    next_data_script = soup.find('script', id='__NEXT_DATA__')

    if next_data_script:
        try:
            props_data = json.loads(next_data_script.string)
            return props_data
        except json.JSONDecodeError:
            print("Error: Could not parse JSON data from script tag")
            return None

    print("Could not find Next.js data in the HTML")
    return None


def import_file(filename: str):
    with open(filename, 'r', encoding='utf-8') as f:
        html_content = f.read()
        props = extract_next_props(html_content)
        if not props:
            return
        tours = props['props']['pageProps']['pack']['tours']
        all_questions = chain.from_iterable(tour['questions'] for tour in tours)
        for question_dict in all_questions:
            question = Question.build_question(question_dict)
            insert_question(question)


if __name__ == '__main__':
    setup_database()
    import_file('test.html')
