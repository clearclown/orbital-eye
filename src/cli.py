"""CLI entry point"""
import typer
from rich.console import Console

app = typer.Typer(name="orbital-eye", help="Satellite Imagery Intelligence Platform")
console = Console()


@app.command()
def detect(
    image: str = typer.Argument(..., help="Path to satellite image (GeoTIFF)"),
    objects: str = typer.Option("all", help="Object types: vessels,aircraft,vehicles,facilities"),
    output: str = typer.Option("output/", help="Output directory"),
    confidence: float = typer.Option(0.5, help="Detection confidence threshold"),
    model: str = typer.Option("auto", help="Detection model to use"),
):
    """Detect objects in a satellite image."""
    console.print(f"[bold green]Detecting objects in {image}...[/]")
    # TODO: implement


@app.command()
def monitor(
    lat: float = typer.Option(..., help="Latitude"),
    lon: float = typer.Option(..., help="Longitude"),
    radius: float = typer.Option(10.0, help="Radius in km"),
    interval: str = typer.Option("7d", help="Check interval"),
    source: str = typer.Option("sentinel2", help="Imagery source"),
):
    """Monitor a location for changes."""
    console.print(f"[bold green]Monitoring ({lat}, {lon})...[/]")
    # TODO: implement


@app.command()
def search(
    lat: float = typer.Option(..., help="Center latitude"),
    lon: float = typer.Option(..., help="Center longitude"),
    radius: float = typer.Option(50.0, help="Search radius in km"),
    target: str = typer.Option("military", help="What to search for: military,naval,airfield,nuclear"),
):
    """Search an area for facilities of interest."""
    console.print(f"[bold green]Searching {radius}km around ({lat}, {lon}) for {target}...[/]")
    # TODO: implement


@app.command()
def report(
    location: str = typer.Option(..., help="Location name or coordinates"),
    period: str = typer.Option("30d", help="Analysis period"),
    output: str = typer.Option("report.html", help="Output file"),
):
    """Generate an intelligence report for a location."""
    console.print(f"[bold green]Generating report for {location}...[/]")
    # TODO: implement


@app.command()
def fetch(
    lat: float = typer.Option(..., help="Center latitude"),
    lon: float = typer.Option(..., help="Center longitude"),
    radius: float = typer.Option(10.0, help="Radius in km"),
    source: str = typer.Option("sentinel2", help="Source: sentinel2, landsat, planet"),
    date_from: str = typer.Option(None, help="Start date (YYYY-MM-DD)"),
    date_to: str = typer.Option(None, help="End date (YYYY-MM-DD)"),
    max_cloud: float = typer.Option(20.0, help="Max cloud cover %"),
):
    """Fetch satellite imagery for an area."""
    console.print(f"[bold green]Fetching imagery from {source}...[/]")
    # TODO: implement


if __name__ == "__main__":
    app()
