import time
from elo import TournamentSystem

ts = TournamentSystem(list(range(10000)), std_dev_multiplier=1.5, initial_phase_matches=5)

start = time.time()


for i in range(50000):
    a, b = ts.select_pair()
    ts.record_winner(winner_id=a, loser_id=b, timestamp=i)

for i, item in enumerate(ts.get_top_items(250), 1):
    print(f'{i}. {item[0]}: {item[1]:.2f} ({item[2]} matches)')
print(ts.get_statistics())
print(time.time() - start)
