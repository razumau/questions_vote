c = get_config()
c.InteractiveShellApp.exec_lines = [
    "from elo import Elo",
    "from models import *",
    "from db import *",
    "from console.helpers import *",
    'print("Project environment loaded!")',
]
