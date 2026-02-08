# Known Facilities Database

Open-source databases of military and strategic facilities.

## Sources
- **SIPRI**: Military expenditure and arms databases (https://www.sipri.org/databases)
- **NTI Nuclear**: Nuclear facilities worldwide (https://www.nti.org/analysis/articles/nuclear-facilities/)
- **OpenStreetMap**: `military=*` tagged features (https://wiki.openstreetmap.org/wiki/Tag:military)
- **Janes**: Open source defence intelligence
- **CSIS Beyond Parallel**: North Korea facility tracking (https://beyondparallel.csis.org/)
- **38 North**: North Korea analysis (https://www.38north.org/)
- **Federation of American Scientists**: Nuclear weapons and facilities

## Format
Each facility entry (JSON):
```json
{
  "id": "facility-001",
  "name": "Yulin Naval Base",
  "type": "naval_base",
  "subtypes": ["submarine_base"],
  "country": "CN",
  "location": {"lat": 18.2269, "lon": 109.5331},
  "sources": ["satellite_imagery", "38north"],
  "last_updated": "2025-01-15"
}
```
