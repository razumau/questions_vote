import apsw
import math
from typing import Union


def create_stddev_function(connection: apsw.Connection) -> None:
    class StdDev:
        def __init__(self):
            self.n: int = 0
            self.sum: float = 0.0
            self.sum_sq: float = 0.0

        def step(self, value: Union[int, float]) -> None:
            if value is not None:
                value = float(value)
                self.n += 1
                self.sum += value
                self.sum_sq += value * value

        def final(self) -> Union[float, None]:
            if self.n < 1:
                return None

            mean = self.sum / self.n
            variance = (self.sum_sq / self.n) - (mean * mean)
            # Handle floating point precision errors that could result in small negative values
            if variance < 0:
                variance = 0
            return math.sqrt(variance)

    connection.create_aggregate_function("stddev", StdDev)
