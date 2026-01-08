"""Backend implementations for search and content fetching."""

from .content_fetcher import (
    AbstractContentFetcher,
    ContentFetcherFactory,
    SourcegraphContentFetcher,
    ZoektContentFetcher,
)
from .models import FormattedResult, Match
from .search import (
    AbstractSearchClient,
    SearchClientFactory,
    SourcegraphSearchClient,
    ZoektSearchClient,
)

__all__ = [
    "AbstractSearchClient",
    "SearchClientFactory",
    "ZoektSearchClient",
    "SourcegraphSearchClient",
    "AbstractContentFetcher",
    "ContentFetcherFactory",
    "ZoektContentFetcher",
    "SourcegraphContentFetcher",
    "FormattedResult",
    "Match",
]

