"""Simple web frontend for the multi-agent system."""

import asyncio
import os
from pathlib import Path

from dotenv import load_dotenv
from fastapi import FastAPI, Form, Request
from fastapi.responses import HTMLResponse
from fastapi.staticfiles import StaticFiles
from fastapi.templating import Jinja2Templates

from servers.context.agent import CodeSnippetFinder, QueryReformater

load_dotenv()

app = FastAPI(title="Multi-Agent Code Search System")

# Templates
templates_dir = Path(__file__).parent / "templates"
templates = Jinja2Templates(directory=str(templates_dir))

# Static files
static_dir = Path(__file__).parent / "static"
if static_dir.exists():
    app.mount("/static", StaticFiles(directory=str(static_dir)), name="static")


@app.get("/", response_class=HTMLResponse)
async def index(request: Request):
    """Main page."""
    return templates.TemplateResponse("index.html", {"request": request})


@app.post("/api/query-reformulate")
async def query_reformulate(question: str = Form(...)):
    """Reformulate user query."""
    try:
        async with QueryReformater() as agent:
            result = await agent.run(question)
        
        return {
            "success": True,
            "original_query": question,
            "suggested_queries": result.suggested_queries,
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
        }


@app.post("/api/code-search")
async def code_search(question: str = Form(...)):
    """Search for code snippets."""
    import traceback
    try:
        # Try using the agent first
        async with CodeSnippetFinder() as agent:
            result = await agent.run(question)
        
        return {
            "success": True,
            "question": question,
            "answer": result,
        }
    except Exception as e:
        # Fallback: Call Zoekt directly
        try:
            import requests
            zoekt_url = os.getenv("ZOEKT_API_URL", "http://localhost:6071")
            response = requests.get(
                f"{zoekt_url}/api/nl-search",
                params={"q": question, "mode": "hybrid", "direct": "false"},
                timeout=30,
            )
            response.raise_for_status()
            data = response.json()
            
            # Format the results
            results = []
            answer_parts = [f"Found {data.get('resultCount', 0)} results for '{question}':"]
            
            # Handle Zoekt results format
            if "results" in data and data["results"]:
                zoekt_result = data["results"]
                if isinstance(zoekt_result, dict) and "Files" in zoekt_result:
                    files = zoekt_result["Files"][:10]  # Limit to 10 results
                    for file_match in files:
                        filename = file_match.get("FileName", "")
                        repo = file_match.get("Repository", "")
                        matches = []
                        if "LineMatches" in file_match:
                            for line_match in file_match["LineMatches"][:5]:  # Limit to 5 matches per file
                                matches.append({
                                    "line_number": line_match.get("LineNum", 0),
                                    "content": str(line_match.get("Line", ""))[:200],
                                })
                        results.append({
                            "repository": repo,
                            "filename": filename,
                            "matches": matches,
                        })
                        answer_parts.append(f"- {filename} (in {repo})")
            
            # Also check semantic results
            if "semanticResults" in data:
                answer_parts.append(f"\nSemantic matches: {len(data['semanticResults'])}")
            
            return {
                "success": True,
                "question": question,
                "answer": "\n".join(answer_parts),
                "results": results,
                "fallback": True,
                "note": "Used direct Zoekt search (agent unavailable)",
            }
        except Exception as fallback_error:
            return {
                "success": False,
                "error": f"Agent failed: {str(e)[:150]}. Zoekt fallback also failed: {str(fallback_error)[:150]}",
            }


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv("FRONTEND_PORT", "3000"))
    uvicorn.run(app, host="0.0.0.0", port=port)

