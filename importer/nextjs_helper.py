import json
from typing import Optional

import requests
from bs4 import BeautifulSoup


def extract_next_props(response: requests.Response) -> Optional[dict]:
    html_content = response.text
    soup = BeautifulSoup(html_content, "html.parser")
    next_data_script = soup.find("script", id="__NEXT_DATA__")

    if next_data_script:
        try:
            props_data = json.loads(next_data_script.string)
            return props_data
        except json.JSONDecodeError:
            print("Error: Could not parse JSON data from script tag")
            return None

    print("Could not find Next.js data in the HTML")
    return None
