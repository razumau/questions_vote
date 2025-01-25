from dataclasses import dataclass
from typing import List, Tuple, Dict

from models import Tournament, TournamentQuestion


@dataclass
class Item:
    id: int
    rating: float = 1500.0
    matches: int = 0
    wins: int = 0


class Elo:
    def __init__(self, tournament: Tournament):
        self.tournament_id = tournament.id
        self.initial_k = tournament.initial_k
        self.minimum_k = tournament.minimum_k
        self.std_dev_multiplier = tournament.std_dev_multiplier
        self.initial_phase_matches = tournament.initial_phase_matches
        self.transition_phase_matches = tournament.transition_phase_matches
        self.top_n = tournament.top_n
        self.band_size = tournament.band_size

    def _calculate_k_factor(self, item: Item) -> float:
        if item.matches < self.initial_phase_matches:
            return self.initial_k
        elif item.matches < self.transition_phase_matches:
            return self.initial_k / 2
        else:
            return self.minimum_k

    def select_pair(self) -> Tuple[int, int]:
        first = TournamentQuestion.get_random_question(
            tournament_id=self.tournament_id, min_rating=self.calculate_threshold()
        )
        while True:
            second = TournamentQuestion.get_random_question(
                tournament_id=self.tournament_id,
                min_rating=first.rating - self.band_size,
                max_rating=first.rating + self.band_size,
            )
            if second.question_id != first.question_id:
                break

        return first, second

    def record_winner(self, winner_id: int, loser_id: int, timestamp: int):
        self.history.append((winner_id, loser_id, timestamp))

        winner = self.items[winner_id]
        loser = self.items[loser_id]

        winner.matches += 1
        loser.matches += 1
        winner.wins += 1

        expected_winner = 1 / (1 + 10 ** ((loser.rating - winner.rating) / 400))
        winner_k = self._calculate_k_factor(winner)
        loser_k = self._calculate_k_factor(loser)
        k_factor = (winner_k + loser_k) / 2

        rating_change = k_factor * (1 - expected_winner)

        winner.rating += rating_change
        loser.rating -= rating_change

    def get_top_items(self, n: int = 100) -> List[Tuple[int, float, int, int]]:
        sorted_items = sorted(self.items.values(), key=lambda x: (x.rating, x.wins / max(1, x.matches)), reverse=True)
        return [(item.id, item.rating, item.matches, item.wins) for item in sorted_items[:n]]

    def calculate_threshold(self) -> float:
        qualified_items = TournamentQuestion.get_qualified_questions(self.tournament_id, self.initial_phase_matches)
        if len(qualified_items) < self.top_n:
            return float("-inf")

        sorted_items = sorted(qualified_items, key=lambda x: x.rating, reverse=True)
        top_n_threshold = sorted_items[min(self.top_n - 1, len(sorted_items) - 1)].rating

        ratings = [item.rating for item in qualified_items]
        mean_rating = sum(ratings) / len(ratings)
        std_dev = (sum((r - mean_rating) ** 2 for r in ratings) / len(ratings)) ** 0.5

        return top_n_threshold - (self.std_dev_multiplier * std_dev)

    def get_statistics(self) -> Dict:
        threshold = self.calculate_threshold()
        viable_items = sum(1 for item in self.items.values() if item.rating >= threshold)

        return {
            "total_comparisons": len(self.history),
            "items_compared": len(self.pairs_seen),
            "average_matches_per_item": sum(i.matches for i in self.items.values()) / len(self.items),
            "rating_range": (min(i.rating for i in self.items.values()), max(i.rating for i in self.items.values())),
            "current_threshold": threshold,
            "viable_items_count": viable_items,
        }
