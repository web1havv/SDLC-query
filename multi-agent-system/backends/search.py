"""Search backends for Zoekt and Sourcegraph."""

from abc import ABC, abstractmethod
from typing import List

from backends.models import FormattedResult


class AbstractSearchClient(ABC):
    """Abstract base class for search clients."""

    @abstractmethod
    def search(self, query: str, num_results: int) -> List[dict]:
        """Search for code.

        Args:
            query: Search query string
            num_results: Maximum number of results

        Returns:
            List of search results
        """
        pass

    @abstractmethod
    def format_results(self, results: List[dict], num_results: int) -> List[FormattedResult]:
        """Format search results.

        Args:
            results: Raw search results
            num_results: Maximum number of results

        Returns:
            List of formatted results
        """
        pass


class ZoektSearchClient(AbstractSearchClient):
    """Zoekt search client implementation."""

    def __init__(self, base_url: str) -> None:
        """Initialize Zoekt client.

        Args:
            base_url: Zoekt API base URL
        """
        self.base_url = base_url

    def search(self, query: str, num_results: int) -> List[dict]:
        """Search using Zoekt."""
        import requests

        # Use /api/nl-search with direct=true to send Zoekt query directly
        # The zoekt-nl-query server accepts queries and returns results
        response = requests.get(
            f"{self.base_url}/api/nl-search",
            params={
                "q": query,
                "mode": "keyword",  # Use keyword mode for direct Zoekt queries
                "direct": "true",   # Skip NL translation, use query as-is
            },
            timeout=30,
        )
        response.raise_for_status()
        result = response.json()
        
        # Extract files from results
        files = []
        if "results" in result and result.get("results"):
            # Results from zoekt-nl-query server are in zoekt.SearchResult format
            # When JSON marshaled, Files becomes a list
            zoekt_result = result["results"]
            if isinstance(zoekt_result, dict) and "Files" in zoekt_result:
                for file_match in zoekt_result["Files"]:
                    matches = []
                    if "LineMatches" in file_match:
                        for line_match in file_match["LineMatches"]:
                            matches.append({
                                "line_number": line_match.get("LineNum", 0),
                                "content": str(line_match.get("Line", "")),
                            })
                    files.append({
                        "repository": file_match.get("Repository", ""),
                        "filename": file_match.get("FileName", ""),
                        "matches": matches,
                    })
            # Alternative: handle if Files is at top level
            elif isinstance(zoekt_result, list):
                files = zoekt_result[:num_results]
        
        return files[:num_results]

    def format_results(self, results: List[dict], num_results: int) -> List[FormattedResult]:
        """Format Zoekt results."""
        from backends.models import FormattedResult, Match

        formatted = []
        for result in results[:num_results]:
            matches = [
                Match(line_number=match.get("line_number", 0), content=match.get("content", ""))
                for match in result.get("matches", [])
            ]
            formatted.append(
                FormattedResult(
                    repository=result.get("repository", ""),
                    filename=result.get("filename", ""),
                    matches=matches,
                )
            )
        return formatted


class SourcegraphSearchClient(AbstractSearchClient):
    """Sourcegraph search client implementation."""

    def __init__(self, endpoint: str, token: str = "") -> None:
        """Initialize Sourcegraph client.

        Args:
            endpoint: Sourcegraph API endpoint
            token: Optional API token
        """
        self.endpoint = endpoint
        self.token = token

    def search(self, query: str, num_results: int) -> List[dict]:
        """Search using Sourcegraph."""
        # Implementation would make HTTP request to Sourcegraph API
        # This is a placeholder
        import requests

        headers = {}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"

        response = requests.post(
            f"{self.endpoint}/.api/graphql",
            json={"query": f'{{ search(query: "{query}", first: {num_results}) {{ results }} }}'},
            headers=headers,
        )
        response.raise_for_status()
        return response.json().get("data", {}).get("search", {}).get("results", [])

    def format_results(self, results: List[dict], num_results: int) -> List[FormattedResult]:
        """Format Sourcegraph results."""
        from backends.models import FormattedResult, Match

        formatted = []
        for result in results[:num_results]:
            matches = [
                Match(line_number=match.get("line_number", 0), content=match.get("content", ""))
                for match in result.get("matches", [])
            ]
            formatted.append(
                FormattedResult(
                    repository=result.get("repository", ""),
                    filename=result.get("filename", ""),
                    matches=matches,
                )
            )
        return formatted


class SearchClientFactory:
    """Factory for creating search clients."""

    @staticmethod
    def create_client(backend: str, **kwargs) -> AbstractSearchClient:
        """Create a search client for the given backend.

        Args:
            backend: Backend name ('zoekt' or 'sourcegraph')
            **kwargs: Backend-specific configuration

        Returns:
            Search client instance

        Raises:
            ValueError: If backend is not supported
        """
        backend = backend.lower()
        if backend == "zoekt":
            return ZoektSearchClient(base_url=kwargs.get("base_url", ""))
        elif backend == "sourcegraph":
            return SourcegraphSearchClient(
                endpoint=kwargs.get("endpoint", ""), token=kwargs.get("token", "")
            )
        else:
            raise ValueError(f"Unsupported backend: {backend}")

