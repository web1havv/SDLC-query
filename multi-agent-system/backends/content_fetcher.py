"""Content fetcher backends for Zoekt and Sourcegraph."""

from abc import ABC, abstractmethod


class AbstractContentFetcher(ABC):
    """Abstract base class for content fetchers."""

    @abstractmethod
    def get_content(self, repo: str, path: str) -> str:
        """Get file or directory content.

        Args:
            repo: Repository path
            path: File or directory path within repository

        Returns:
            File content or directory listing

        Raises:
            ValueError: If path or repository doesn't exist
        """
        pass


class ZoektContentFetcher(AbstractContentFetcher):
    """Zoekt content fetcher implementation."""

    def __init__(self, zoekt_url: str) -> None:
        """Initialize Zoekt content fetcher.

        Args:
            zoekt_url: Zoekt API base URL
        """
        self.zoekt_url = zoekt_url

    def get_content(self, repo: str, path: str) -> str:
        """Get content from Zoekt.
        
        Note: The zoekt-nl-query server doesn't expose a content API endpoint yet.
        This is a placeholder that would need to be implemented if needed.
        """
        import requests

        # For now, return a message that this endpoint isn't available
        # In the future, you could add a content endpoint to zoekt-nl-query server
        # or use Zoekt's native file content API if available
        try:
            # Try to fetch from a potential content endpoint (not implemented yet)
            response = requests.get(
                f"{self.zoekt_url}/api/content",
                params={"repo": repo, "path": path},
                timeout=10,
            )
            response.raise_for_status()
            return response.text
        except requests.exceptions.RequestException:
            # Content fetching not yet implemented in zoekt-nl-query server
            return f"Content fetching for {repo}/{path} is not yet implemented. " \
                   f"Please use the zoekt-nl-query dashboard at {self.zoekt_url}/dashboard for file browsing."


class SourcegraphContentFetcher(AbstractContentFetcher):
    """Sourcegraph content fetcher implementation."""

    def __init__(self, endpoint: str, token: str = "") -> None:
        """Initialize Sourcegraph content fetcher.

        Args:
            endpoint: Sourcegraph API endpoint
            token: Optional API token
        """
        self.endpoint = endpoint
        self.token = token

    def get_content(self, repo: str, path: str) -> str:
        """Get content from Sourcegraph."""
        import requests

        headers = {}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"

        response = requests.get(
            f"{self.endpoint}/.api/repos/{repo}/file/{path}",
            headers=headers,
        )
        response.raise_for_status()
        return response.text


class ContentFetcherFactory:
    """Factory for creating content fetchers."""

    @staticmethod
    def create_fetcher(backend: str, **kwargs) -> AbstractContentFetcher:
        """Create a content fetcher for the given backend.

        Args:
            backend: Backend name ('zoekt' or 'sourcegraph')
            **kwargs: Backend-specific configuration

        Returns:
            Content fetcher instance

        Raises:
            ValueError: If backend is not supported
        """
        backend = backend.lower()
        if backend == "zoekt":
            return ZoektContentFetcher(zoekt_url=kwargs.get("zoekt_url", ""))
        elif backend == "sourcegraph":
            return SourcegraphContentFetcher(
                endpoint=kwargs.get("endpoint", ""), token=kwargs.get("token", "")
            )
        else:
            raise ValueError(f"Unsupported backend: {backend}")

