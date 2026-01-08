"""Models for evaluator components."""

from pydantic import BaseModel, Field


class CodeSnippetResult(BaseModel):
    """Result containing code snippet information."""

    code: str = Field(..., description="sample code snippet")
    language: str = Field(..., description="language of the code snippet")
    description: str = Field(..., description="description of the code snippet")

