"""Backend models for search and content fetching."""

from dataclasses import dataclass
from typing import List


@dataclass
class Match:
    """Represents a match in search results."""

    line_number: int
    content: str = ""


@dataclass
class FormattedResult:
    """Formatted search result."""

    repository: str
    filename: str
    matches: List[Match]

