"""REST API server"""
from fastapi import FastAPI
app = FastAPI(title="Orbital Eye", version="0.1.0")

@app.get("/health")
def health():
    return {"status": "ok"}

# TODO: implement endpoints
