import os

from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import Application, CommandHandler, CallbackQueryHandler, ContextTypes

from db import connection, setup_database
from elo import Elo
from models import Tournament


async def list_active_tournaments(update: Update, context: ContextTypes.DEFAULT_TYPE):
    tournaments = Tournament.list_active_tournaments()
    tournaments_str = "\n".join([f"{t.id}: {t.title}" for t in tournaments])
    message = f"Active tournaments: \n{tournaments_str}"
    await context.bot.send_message(chat_id=update.effective_chat.id, text=message)

async def stats_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    t = Tournament.find_active_tournament()
    elo = Elo(t)
    stats = elo.get_statistics()
    message = f"""
        Threshold: {stats['current_threshold']}.
        There are {stats['above_threshold']} questions above the threshold.
        {stats['unqualified']} questions are unqualified.
        Rating distribution: {stats['distribution']}. 
        Total matches: {stats['total_matches']}, total wins: {stats['total_wins']}.
    """
    await context.bot.send_message(chat_id=update.effective_chat.id, text=message)

def main():
    setup_database()
    application = Application.builder().token(os.getenv("ADMIN_TELEGRAM_TOKEN")).build()

    application.add_handler(CommandHandler("stats", stats_command))
    application.add_handler(CommandHandler("active_tournaments", list_active_tournaments))

    application.run_polling()


if __name__ == "__main__":
    main()