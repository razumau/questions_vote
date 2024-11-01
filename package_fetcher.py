import requests
from typing import Optional
from urllib.parse import urlparse

def fetch_package(package_id: int) -> Optional[str]:
    url = f"https://gotquestions.online/pack/{package_id}"
    try:
        return fetch_html(url)

    except HTMLFetchError as e:
        print(f"Error: {str(e)}")
        return None
    except Exception as e:
        print(f"Unexpected error: {str(e)}")
        return None


class HTMLFetchError(Exception):
    pass


def fetch_html(url: str) -> str:
    try:
        result = urlparse(url)
        if not all([result.scheme, result.netloc]):
            raise ValueError("Invalid URL format")
    except Exception as e:
        raise HTMLFetchError(f"Invalid URL: {str(e)}")

    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8',
        'Accept-Language': 'en-US,en;q=0.5',
    }

    try:
        response = requests.get(url, headers=headers, timeout=10)
        response.raise_for_status()

        content_type = response.headers.get('content-type', '').lower()
        if 'text/html' not in content_type:
            raise HTMLFetchError(f"Unexpected content type: {content_type}")

        response.encoding = response.apparent_encoding
        return response.text

    except requests.RequestException as e:
        raise HTMLFetchError(f"Error fetching URL: {str(e)}")
