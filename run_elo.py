import time
import cProfile
import pstats
from elo import Elo
from models import Tournament, Question, TournamentQuestion
from db import setup_database


# Tournament.create_tournament(
#     question_ids=Question.question_ids_for_year(2022),
#     title="Test Tournament",
#     initial_k=64.0,
#     minimum_k=16.0,
#     std_dev_multiplier=1.5,
#     initial_phase_matches=5,
#     transition_phase_matches=10,
#     top_n=100,
#     band_size=200
# )

# Tournament.start_tournament("Test Tournament")


def main():
    start = time.time()

    setup_database()

    t = Tournament.find_active_tournament()
    ts = Elo(t)

    print(TournamentQuestion.get_rating_distribution(t.id))

    for i in range(1000):
        a, b = ts.select_pair()
        ts.record_winner(winner_id=a, loser_id=b)
        if i % 100 == 0:
            print(f"Pair {i}")
            print(ts.get_statistics())

    print(ts.get_statistics())

    duration = time.time() - start
    print(f"{duration:.2f} seconds")


if __name__ == "__main__":
    profiler = cProfile.Profile()
    profiler.enable()
    main()
    profiler.disable()
    profiler.dump_stats("profile_results.prof")
    stats = pstats.Stats(profiler).sort_stats("cumtime")
    stats.print_stats()
