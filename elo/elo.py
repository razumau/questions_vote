from models import Tournament, TournamentQuestion

_retries = 0


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

    def _select_two_unqualified(self) -> tuple[int, int]:
        first = TournamentQuestion.get_random_question(self.tournament_id, max_matches=self.initial_phase_matches - 1)
        second = TournamentQuestion.get_random_question(self.tournament_id, max_matches=self.initial_phase_matches - 1)
        return first.question_id, second.question_id

    def _select_two_qualified(self) -> tuple[int, int]:
        threshold = self.calculate_threshold()
        first = TournamentQuestion.get_random_question(tournament_id=self.tournament_id, min_rating=threshold)
        second = TournamentQuestion.get_random_question(tournament_id=self.tournament_id, min_rating=threshold)
        return first.question_id, second.question_id

    def select_pair(self) -> tuple[int, int]:
        global _retries
        unqualified_count = TournamentQuestion.count_unqualified_questions(
            self.tournament_id, self.initial_phase_matches
        )
        if unqualified_count > 1:
            first, second = self._select_two_unqualified()
        elif unqualified_count == 1:
            first = TournamentQuestion.get_random_question(
                self.tournament_id, max_matches=self.initial_phase_matches - 1
            )
            second = TournamentQuestion.get_random_question(self.tournament_id)
        else:
            first, second = self._select_two_qualified()

        if first == second:
            _retries += 1
            return self.select_pair()

        return first, second

    def _calculate_k_factor(self, item: TournamentQuestion) -> float:
        if item.matches < self.initial_phase_matches:
            return self.initial_k
        elif item.matches < self.transition_phase_matches:
            return self.initial_k / 2
        else:
            return self.minimum_k

    def record_winner(self, winner_id: int, loser_id: int) -> None:
        winner = TournamentQuestion.find(self.tournament_id, winner_id)
        loser = TournamentQuestion.find(self.tournament_id, loser_id)

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

        winner.save()
        loser.save()

    def get_top_items(self, n: int = 100) -> list:
        return TournamentQuestion.get_top_questions(self.tournament_id, n)

    def calculate_threshold(self) -> float:
        ratings_count, std_dev = TournamentQuestion.get_stats_for_qualified(
            self.tournament_id, self.initial_phase_matches
        )
        if ratings_count < self.top_n:
            return float("-inf")

        top_n_threshold = TournamentQuestion.get_rating_at_position(self.tournament_id, self.top_n)
        return top_n_threshold - (self.std_dev_multiplier * std_dev)

    def get_questions_stats(self, q1_id: int, q2_id: int) -> list[dict]:
        q1 = TournamentQuestion.find(self.tournament_id, q1_id)
        q2 = TournamentQuestion.find(self.tournament_id, q2_id)
        return [{"matches": q1.matches, "wins": q1.wins}, {"matches": q2.matches, "wins": q2.wins}]

    def get_statistics(self) -> dict:
        global _retries
        threshold = self.calculate_threshold()
        above_threshold_count = TournamentQuestion.count_questions_above_threshold(self.tournament_id, threshold)
        rating_distribution = TournamentQuestion.get_rating_distribution(self.tournament_id)
        unqualified_count = TournamentQuestion.count_unqualified_questions(
            self.tournament_id, self.initial_phase_matches
        )
        match_counts = TournamentQuestion.get_match_counts(self.tournament_id)

        return {
            "current_threshold": threshold,
            "above_threshold": above_threshold_count,
            "unqualified": unqualified_count,
            "distribution": rating_distribution,
            "retries": _retries,
            "total_matches": match_counts["matches"],
            "total_wins": match_counts["wins"],
        }
