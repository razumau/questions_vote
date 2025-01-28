import logging
import os

from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import Application, CommandHandler, CallbackQueryHandler, ContextTypes

from db import connection, setup_database
from elo import Elo
from models import Tournament, Question

# Set up logging
logging.basicConfig(format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.INFO)


def save_vote(user_id: int, question1_id: int, question2_id: int, selected_id: int = None):
    with connection() as conn:
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
    keyboard = [
        [
            InlineKeyboardButton("Question 1", callback_data=f"vote_{q1_id}_{q2_id}_1"),
            InlineKeyboardButton("Question 2", callback_data=f"vote_{q1_id}_{q2_id}_2"),
        ],
        [InlineKeyboardButton("Neither", callback_data=f"vote_{q1_id}_{q2_id}_0")],
    ]
    return InlineKeyboardMarkup(keyboard)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    await update.message.reply_text("Welcome to the Question Vote Bot! Use /vote to get two random questions to compare.")

def get_questions():
    t = Tournament.find_active_tournament()
    ts = Elo(t)
    q1_id, q2_id = ts.select_pair()
    return Question.find([q1_id, q2_id])

async def vote_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    q1, q2 = get_questions()
    if q1['image_data']:
        await context.bot.send_photo(chat_id=update.effective_chat.id, photo=q1['image_data'], caption='К вопросу 1')

    if q2['image_data']:
        await context.bot.send_photo(chat_id=update.effective_chat.id, photo=q2['image_data'], caption='К вопросу 2')
    message = f"Which question is better?\n\nQuestion 1:\n{q1['question']}\n\nQuestion 2:\n{q2['question']}"
    keyboard = create_vote_keyboard(q1['id'], q2['id'])
    await context.bot.send_message(chat_id=update.effective_chat.id, text=message, reply_markup=keyboard)


async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE):
    query = update.callback_query
    await query.answer()

    _, q1_id, q2_id, choice = query.data.split("_")
    choice = int(choice)

    selected_id = None
    if choice == 1:
        selected_id = int(q1_id)
        response = "You voted for Question 1!"
    elif choice == 2:
        selected_id = int(q2_id)
        response = "You voted for Question 2!"
    else:
        response = "You voted for neither question."

    save_vote(user_id=query.from_user.id, question1_id=int(q1_id), question2_id=int(q2_id), selected_id=selected_id)

    await query.edit_message_text(text=response)
    await vote_command(update, context)


def main():
    setup_database()
    application = Application.builder().token(os.getenv("TELEGRAM_TOKEN")).build()

    application.add_handler(CommandHandler("start", start))
    application.add_handler(CommandHandler("vote", vote_command))
    application.add_handler(CallbackQueryHandler(button_callback))

    application.run_polling()


if __name__ == "__main__":
    main()
