import logging
import os
import textwrap
from functools import lru_cache
from typing import Optional

import sentry_sdk
from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup, LinkPreviewOptions
from telegram.constants import ParseMode
from telegram.ext import Application, CommandHandler, CallbackQueryHandler, ContextTypes

from db import setup_database
from elo import Elo
from models import Tournament, Question, Vote
from rate_limiter import RateLimiter


logging.basicConfig(format="%(asctime)s - %(name)s - %(levelname)s - %(message)s", level=logging.WARN)

if os.getenv("SENTRY_DSN"):
    sentry_sdk.init(
        dsn=os.getenv("SENTRY_DSN"),
        send_default_pii=True,
        traces_sample_rate=1.0,
        profiles_sample_rate=1.0,
    )


@lru_cache(maxsize=1)
def questions_count():
    return Tournament.find_active_tournament().questions_count


@lru_cache(maxsize=1)
def elo():
    t = Tournament.find_active_tournament()
    return Elo(t)


def get_questions() -> list[Question]:
    q1_id, q2_id = elo().select_pair()
    return Question.find([q1_id, q2_id])


@sentry_sdk.trace
def save_vote(user_id: int, question1_id: int, question2_id: int, selected_id: Optional[int]):
    Vote.create(user_id, question1_id, question2_id, elo().tournament_id, selected_id)
    if selected_id is None:
        return

    loser_id = question1_id if selected_id == question2_id else question2_id
    elo().record_winner(winner_id=selected_id, loser_id=loser_id)


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
    await update.message.reply_text(
        f"Вспомнить хорошие вопросы спустя год сложно, поэтому я предлагаю вам сравнить все "
        f"вопросы за 2022 год ({questions_count()} штук). Бот будет присылать пары вопросов, и вам нужно будет выбрать, "
        f"какой из двух лучше. Цель этого этапа — выбрать шортлист из 100 вопросов. Вопросы с низким (Эло-подобным) "
        f"рейтингом будут постепенно убираться, но не раньше, чем каждый поучаствует в пяти матчах."
    )


def format_question(question: Question, number: int) -> str:
    handout = f"<b>Раздаточный материал</b>:\n{question.handout_str}\n\n" if question.handout_str else ""
    accepted = f"\n<b>Зачёт</b>: {question.accepted_answer}" if question.accepted_answer else ""

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


def confirmation_message(q1_id: int, q2_id: int, selected_id: int) -> str:
    if selected_id == q1_id:
        return "Записали, что первый вопрос лучше."
    elif selected_id == q2_id:
        return "Записали, что второй вопрос лучше."
    else:
        return "Ок, сейчас пришлём другую пару вопросов."


def inflect_wins(number: int) -> str:
    if 11 <= number % 100 <= 19:
        wins_word = "побед"
    elif number % 10 == 1:
        wins_word = "победа"
    elif 2 <= number % 10 <= 4:
        wins_word = "победы"
    else:
        wins_word = "побед"
    return f"{number} {wins_word}"


def inflect_matches(number: int) -> str:
    if number == 1:
        matches_word = "матче"
    elif number == 0:
        matches_word = "матчей"
    else:
        matches_word = "матчах"
    return f"{number} {matches_word}"


@sentry_sdk.trace
def questions_stats_message(q1_id: int, q2_id: int, selected_id: Optional[int]) -> str:
    if not selected_id:
        return ""

    questions_stats = elo().get_questions_stats(q1_id, q2_id)
    first_wins, first_matches = questions_stats[0]["wins"], questions_stats[0]["matches"]
    second_wins, second_matches = questions_stats[1]["wins"], questions_stats[1]["matches"]
    first_pct = round(first_wins / first_matches * 100, 1) if first_matches else 0
    second_pct = round(second_wins / second_matches * 100, 1) if second_matches else 0

    return (
        f"У первого теперь {inflect_wins(first_wins)} в {inflect_matches(first_matches)} ({first_pct}%), "
        f"а у второго — {inflect_wins(second_wins)} в {inflect_matches(second_matches)} ({second_pct}%)."
    )


@sentry_sdk.trace
async def send_question(chat_id: int, context: ContextTypes.DEFAULT_TYPE, question: Question, number: int):
    if question.handout_img:
        await context.bot.send_photo(chat_id=chat_id, photo=question.handout_img)

    question_text = format_question(question, number)
    await context.bot.send_message(
        chat_id=chat_id,
        text=question_text,
        link_preview_options=LinkPreviewOptions(is_disabled=True),
        parse_mode=ParseMode.HTML,
    )


async def send_vote_job(context: ContextTypes.DEFAULT_TYPE):
    with sentry_sdk.start_transaction(op="task", name="Send vote") as sentry_transaction:
        chat_id = context.job.chat_id
        q1, q2 = get_questions()

        sentry_transaction.set_data("q1_id", q1.id)
        sentry_transaction.set_data("q2_id", q2.id)

        await send_question(chat_id, context, q1, 1)
        await send_question(chat_id, context, q2, 2)

        keyboard = create_vote_keyboard(q1.id, q2.id)
        RateLimiter().record(chat_id)
        await context.bot.send_message(chat_id=chat_id, text="Какой вопрос лучше?", reply_markup=keyboard)


async def vote_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    chat_id = update.effective_chat.id
    when = RateLimiter().can_send_in_seconds(chat_id)
    context.job_queue.run_once(send_vote_job, when=when, chat_id=chat_id, job_kwargs={"misfire_grace_time": None})


async def button_callback(update: Update, context: ContextTypes.DEFAULT_TYPE):
    with sentry_sdk.start_transaction(op="task", name="Record vote"):
        query = update.callback_query
        await query.answer()

        _, q1_id, q2_id, choice = query.data.split("_")
        q1_id, q2_id, choice = int(q1_id), int(q2_id), int(choice)
        selected_id = q1_id if choice == 1 else q2_id if choice == 2 else None

        save_vote(user_id=query.from_user.id, question1_id=q1_id, question2_id=q2_id, selected_id=selected_id)

    response = f"{confirmation_message(q1_id, q2_id, selected_id)} {questions_stats_message(q1_id, q2_id, selected_id)}"
    await query.edit_message_text(text=textwrap.dedent(response))
    await vote_command(update, context)


def main():
    setup_database()
    application = Application.builder().token(os.getenv("TELEGRAM_TOKEN")).build()

    application.add_handler(CommandHandler("start", start, block=False))
    application.add_handler(CommandHandler("vote", vote_command, block=False))
    application.add_handler(CallbackQueryHandler(button_callback, block=False))

    application.run_polling()


if __name__ == "__main__":
    main()
