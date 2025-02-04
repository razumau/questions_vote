from elo import Elo
from models import Tournament, Question


def create_tournament_for_year(year: int):
    title = f"All {year} questions"
    questions = Question.question_ids_for_year(year)
    Tournament.create_tournament(
        question_ids=questions,
        title=title,
        initial_k=64.0,
        minimum_k=16.0,
        std_dev_multiplier=1.5,
        initial_phase_matches=5,
        transition_phase_matches=10,
        top_n=100,
        band_size=200
    )

    Tournament.start_tournament(title)

def active_elo():
    return Elo(Tournament.find_active_tournament())
