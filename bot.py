import logging
import os
import textwrap

from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup, LinkPreviewOptions
from telegram.constants import ParseMode
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
            InlineKeyboardButton("Первый", callback_data=f"vote_{q1_id}_{q2_id}_1"),
            InlineKeyboardButton("Второй", callback_data=f"vote_{q1_id}_{q2_id}_2"),
        ],
        [InlineKeyboardButton("Не могу выбрать", callback_data=f"vote_{q1_id}_{q2_id}_0")],
    ]
    return InlineKeyboardMarkup(keyboard)


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    await update.message.reply_text("Welcome to the Question Vote Bot! Use /vote to get two random questions to compare.")

def format_question(question: Question, number: int) -> str:
    handout = f"<b>Раздаточный материал</b>:\n{question.handout_str}\n\n" if question.handout_str else ""
    accepted = f"<b>Зачёт</b>: {question.accepted_answer}\n" if question.accepted_answer else ""
    formatted = f"""\
    <b>Вопрос {number}.</b>
    {handout}{question.question}
    
    <tg-spoiler>
    <b>Ответ</b>: {question.answer}{accepted}
    <b>Комментарий</b>: {question.comment}
    <b>Источник</b>: {question.source}
    </tg-spoiler>
    """

    return formatted.replace("\n    ", "\n")

def get_questions() -> list[Question]:
    t = Tournament.find_active_tournament()
    ts = Elo(t)
    q1_id, q2_id = ts.select_pair()
    return Question.find([q1_id, q2_id])

async def send_question(update: Update, context: ContextTypes.DEFAULT_TYPE, text):
    await context.bot.send_message(chat_id=update.effective_chat.id, text=text,
                                   link_preview_options=LinkPreviewOptions(is_disabled=True),
                                   parse_mode=ParseMode.HTML)

async def vote_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    q1, q2 = get_questions()
    if q1.handout_img:
        await context.bot.send_photo(chat_id=update.effective_chat.id, photo=q1.handout_img, caption='К вопросу 1')

    if q2.handout_img:
        await context.bot.send_photo(chat_id=update.effective_chat.id, photo=q2.handout_img, caption='К вопросу 2')

    q1_str = format_question(q1, number=1)
    await send_question(update, context, q1_str)

    q2_str = format_question(q2, number=2)
    await send_question(update, context, q2_str)

    keyboard = create_vote_keyboard(q1.id, q2.id)
    await context.bot.send_message(chat_id=update.effective_chat.id, text="Какой вопрос лучше?", reply_markup=keyboard)


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
