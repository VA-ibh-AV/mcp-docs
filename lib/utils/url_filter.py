import re
from urllib.parse import urlparse

FORBIDDEN_EXTENSIONS = [
    "png", "jpg", "jpeg", "gif", "svg", "webp",
    "css", "js", "ico", "map",
    "pdf", "zip", "tar", "gz", "7z", "rar",
    "mp4", "mp3", "wav", "avi", "mov",
    "woff", "woff2", "ttf", "eot",
]

FORBIDDEN_KEYWORDS = [
    "pricing", "plans", "enterprise", "customers", "case-study",
    "about", "company", "team", "careers", "jobs", "hiring",
    "security", "legal", "privacy", "terms", "gdpr",

    "blog", "newsletter", "community", "event", "events",
    "news", "press", "updates", "release-notes", "releases",
    "changelog", "roadmap",

    "store", "shop", "buy", "subscribe", "credits",

    "contact", "support", "help-center", "faq",

    "github.com", "gitlab.com", "bitbucket.org",

    "login", "logout", "signin", "signup", "register", "auth",

    "utm_", "ref=", "source=",

    "print-pdf", "pdf-version", "download",
]

def is_relevant_docs_url(url: str, allowed_domain: str) -> bool:

    try:
        parsed = urlparse(url)
    except Exception:
        return False

    domain = parsed.netloc.lower()

    # Extract domain from allowed_domain (handle if it's a full URL)
    allowed_parsed = urlparse(allowed_domain)
    allowed_netloc = allowed_parsed.netloc.lower() if allowed_parsed.netloc else allowed_domain.lower()

    # Must match allowed domain
    if domain not in {allowed_netloc, f"www.{allowed_netloc}"}:
        return False

    path = parsed.path.lower()

    if path in ["", "/"]:
        return False

    # Block assets/media
    if any(path.endswith(f".{ext}") for ext in FORBIDDEN_EXTENSIONS):
        return False

    # Block forbidden words anywhere
    for keyword in FORBIDDEN_KEYWORDS:
        if keyword in url.lower():
            return False

    return True
