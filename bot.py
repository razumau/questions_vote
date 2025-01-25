import logging
import os

from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import Application, CommandHandler, CallbackQueryHandler, ContextTypes
import sqlite3
from typing import Tuple

from db import DB_PATH

# Set up logging
logging.basicConfig(format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO)


class QuizVoteBot:
    def __init__(self, db_path: str):
        self.db_path = db_path

    def get_random_questions(self) -> Tuple[Tuple[int, str], Tuple[int, str]]:
        with sqlite3.connect(self.db_path) as conn:
            cursor = conn.cursor()
            cursor.execute("""
                            SELECT id, question
                            FROM questions 
                            ORDER BY RANDOM() 
                            LIMIT 2
                        """)

            questions = cursor.fetchall()
            return tuple(questions)

    def save_vote(self, user_id: int, question1_id: int, question2_id: int, selected_id: int = None):
        """Save a user's vote to the database."""
        with sqlite3.connect(self.db_path) as conn:
            cursor = conn.cursor()
            cursor.execute(
                """
                INSERT INTO votes (user_id, question1_id, question2_id, selected_id)
                VALUES (?, ?, ?, ?)
            """,
                (user_id, question1_id, question2_id, selected_id),
            )
            conn.commit()


def create_vote_keyboard(q1_id: int, q2_id: int) -> InlineKeyboardMarkup:
    """Create an inline keyboard for voting."""
    keyboard = [
        [
            InlineKeyboardButton("Question 1", callback_data=f"vote_{q1_id}_{q2_id}_1"),
            InlineKeyboardButton("Question 2", callback_data=f"vote_{q1_id}_{q2_id}_2"),
        ],
        [InlineKeyboardButton("Neither", callback_data=f"vote_{q1_id}_{q2_id}_0")],
    ]
    return InlineKeyboardMarkup(keyboard)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Send a message when the command /start is issued."""
    await update.message.reply_text("Welcome to the Quiz Vote Bot! Use /vote to get two random questions to compare.")


async def vote_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Send two random questions when the command /vote is issued."""
    bot = QuizVoteBot(DB_PATH)
    q1, q2 = bot.get_random_questions()

    message = f"Please vote for your preferred question:\n\nQuestion 1:\n{q1[1]}\n\nQuestion 2:\n{q2[1]}"

    keyboard = create_vote_keyboard(q1[0], q2[0])
    await update.message.reply_text(message, reply_markup=keyboard)


async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle button presses."""
    query = update.callback_query
    await query.answer()

    # Parse callback data
    _, q1_id, q2_id, choice = query.data.split("_")
    choice = int(choice)

    # Save the vote
    bot = QuizVoteBot(DB_PATH)
    selected_id = None
    if choice == 1:
        selected_id = int(q1_id)
        response = "You voted for Question 1!"
    elif choice == 2:
        selected_id = int(q2_id)
        response = "You voted for Question 2!"
    else:
        response = "You voted for neither question."

    bot.save_vote(user_id=query.from_user.id, question1_id=int(q1_id), question2_id=int(q2_id), selected_id=selected_id)

    # First, update the message with the vote result
    await query.edit_message_text(text=response)

    # Then send a new pair of questions
    q1, q2 = bot.get_random_questions()
    message = f"Please vote for your preferred question:\n\nQuestion 1:\n{q1[1]}\n\nQuestion 2:\n{q2[1]}"
    keyboard = create_vote_keyboard(q1[0], q2[0])
    await query.message.reply_text(message, reply_markup=keyboard)


def main():
    application = Application.builder().token(os.getenv("TELEGRAM_TOKEN")).build()

    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("vote", vote_command))
    application.add_handler(CallbackQueryHandler(button_callback))

    application.run_polling()


if __name__ == "__main__":
    main()
