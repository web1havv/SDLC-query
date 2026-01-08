"""Demo script for evaluator components."""

import argparse
import asyncio
import json
import os
import pathlib
import uuid

from pydantic import BaseModel, Field
from pydantic_ai.agent import Agent
from pydantic_ai.models import Model
from pydantic_ai.models.openai import OpenAIModel, OpenAIModelSettings
from pydantic_ai.providers.openai import OpenAIProvider
from pydantic_ai.settings import ModelSettings

from core import PromptManager
from evaluator.config import JudgeConfig
from evaluator.models import CodeSnippetResult
from servers.context.agent import CodeSnippetFinder


async def async_main():
    """Main async function for testing."""
    question = input("Enter a question: ")

    async with CodeSnippetFinder(trace_id=str(uuid.uuid4())) as agent:
        result = await agent.run(question)

    print(f"{'-' * 10}\n{result}\n {'-' * 10}\n")

    with open("question.json", "w") as f:
        json.dump(
            {
                "input": {
                    "question": question,
                },
                "answer": result,
            },
            f,
            indent=2,
            ensure_ascii=True,
        )


if __name__ == "__main__":
    asyncio.run(async_main())

