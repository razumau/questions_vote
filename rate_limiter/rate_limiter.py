from datetime import datetime


class RateLimiter:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance.chats = {}
            cls._instance.timeout = 10
        return cls._instance

    def record(self, chat_id: int) -> None:
        self.chats[chat_id] = datetime.now().timestamp()

    def can_send_in_seconds(self, chat_id: int) -> float:
        last_timestamp = self.chats.get(chat_id)
        if last_timestamp is None:
            return 0

        timestamp = datetime.now().timestamp()
        return self.timeout - (timestamp - last_timestamp)
