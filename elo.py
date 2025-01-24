from dataclasses import dataclass
from typing import List, Tuple, Dict, Set
import random
import math

from models import Tournament


@dataclass
class Item:
    id: int
    rating: float = 1500.0
    matches: int = 0
    wins: int = 0


class Elo:
    def __init__(self, tournament: Tournament):
        self.items = {id: Item(id=id) for id in items}
        self.initial_k = tournament.initial_k
        self.minimum_k = tournament.minimum_k
        self.std_dev_multiplier = tournament.std_dev_multiplier
        self.initial_phase_matches = tournament.initial_phase_matches
        self.transition_phase_matches = tournament.transition_phase_matches
        self.top_n = tournament.top_n
        self.history: List[Tuple[int, int, int]] = []
        self.pairs_seen: Set[Tuple[int, int]] = set()

    def _calculate_k_factor(self, item: Item) -> float:
        if item.matches < self.initial_phase_matches:
            return self.initial_k
        elif item.matches < self.transition_phase_matches:
            return self.initial_k / 2
        else:
            return self.minimum_k

    def select_pair(self) -> Tuple[int, int]:
        if len(self.history) < len(self.items) * 2:
            return self._random_pair()
        else:
            return self._rating_based_pair()

    def _random_pair(self) -> Tuple[int, int]:
        while True:
            pair = tuple(random.sample(list(self.items.keys()), 2))
            if pair not in self.pairs_seen and (pair[1], pair[0]) not in self.pairs_seen:
                self.pairs_seen.add(pair)
                return pair

    def _rating_based_pair(self) -> Tuple[int, int]:
        threshold = self.calculate_threshold()
        viable_items = [item for item in self.items.values()
                        if item.rating >= threshold or item.matches < self.initial_phase_matches]

        if len(viable_items) < 2:
            return self._random_pair()

        band_size = 200
        bands = {}
        for item in viable_items:
            band = math.floor(item.rating / band_size)
            if band not in bands:
                bands[band] = []
            bands[band].append(item)

        selected_band = max(bands.items(), key=lambda x: len(x[1]))[1]

        selected_band.sort(key=lambda x: x.matches)
        item1 = selected_band[0]

        candidates = [i for i in self.items.values()
                      if i.id != item1.id
                      and (i.id, item1.id) not in self.pairs_seen
                      and (item1.id, i.id) not in self.pairs_seen]

        if not candidates:
            return self._random_pair()

        item2 = min(candidates,
                    key=lambda x: abs(x.rating - item1.rating))

        self.pairs_seen.add((item1.id, item2.id))
        return item1.id, item2.id

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
        sorted_items = sorted(self.items.values(),
                              key=lambda x: (x.rating, x.wins / max(1, x.matches)),
                              reverse=True)
        return [(item.id, item.rating, item.matches, item.wins) for item in sorted_items[:n]]

    def calculate_threshold(self) -> float:
        qualified_items = [item for item in self.items.values() if item.matches >= self.initial_phase_matches]
        if len(qualified_items) < self.top_n:
            return float('-inf')

        sorted_items = sorted(qualified_items, key=lambda x: x.rating, reverse=True)
        top_n_threshold = sorted_items[min(self.top_n - 1, len(sorted_items) - 1)].rating

        ratings = [item.rating for item in qualified_items]
        mean_rating = sum(ratings) / len(ratings)
        std_dev = (sum((r - mean_rating) ** 2 for r in ratings) / len(ratings)) ** 0.5

        return top_n_threshold - (self.std_dev_multiplier * std_dev)

    def get_statistics(self) -> Dict:
        threshold = self.calculate_threshold()
        viable_items = sum(1 for item in self.items.values()
                           if item.rating >= threshold)

        return {
            'total_comparisons': len(self.history),
            'items_compared': len(self.pairs_seen),
            'average_matches_per_item': sum(i.matches for i in self.items.values()) / len(self.items),
            'rating_range': (
                min(i.rating for i in self.items.values()),
                max(i.rating for i in self.items.values())
            ),
            'current_threshold': threshold,
            'viable_items_count': viable_items
        }
