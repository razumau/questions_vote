import time
from elo import Elo
from models import Tournament, Question
from db import setup_database
from models.tournament_question import TournamentQuestion

start = time.time()

setup_database()

question_ids = Question.question_ids_for_year(2022)
print(len(question_ids))
print(question_ids[0])


#
# Tournament.create_tournament(
#     question_ids=question_ids,
#     title="Test Tournament",
#     initial_k=64.0,
#     minimum_k=16.0,
#     std_dev_multiplier=1.5,
#     initial_phase_matches=5,
#     transition_phase_matches=10,
#     top_n=100,
#     band_size=200
# )

Tournament.start_tournament("Test Tournament")
t = Tournament.find_active_tournament()

print(t)
print(TournamentQuestion.get_random_question(tournament_id=t.id))

ts = Elo(Tournament.find_active_tournament())

print(ts.select_pair())


# for i in range(50000):
#     a, b = ts.select_pair()
#     ts.record_winner(winner_id=a, loser_id=b, timestamp=i)
#     if i % 1000 == 0:
#         print(f'{i}: {ts.get_statistics()['viable_items_count']}')
#
# for i, item in enumerate(ts.get_top_items(10000), 1):
#     print(f'{i}. {item[0]}: {item[1]:.2f} ({item[2]} matches, {item[3]} wins)')
# print(ts.get_statistics())

print(time.time() - start)
