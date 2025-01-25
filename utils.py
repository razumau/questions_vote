import random
import time


def sleep_around(seconds: float, deviation: float = 0.3):
    time.sleep(random.uniform(seconds - deviation, seconds + deviation))
