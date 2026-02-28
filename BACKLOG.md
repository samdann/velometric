# Velometric Backlog


## New Feature
- [x] My activities: should be an inbox (table) displaying sport type, title, date, time, distance, elevation, avg power | item should be paginated, and page should show 10, 25 or 50 results.
- [x] Overview tab, under elevation graph, build a graph for speed profile (speed by distance). The graph should have the same characteristics as the elevation graph.
- [x] Under overview tab, under the summary, build a graph for display elevation profile (elevation by distance)
- [x] connect swagger-ui to at port 8081, so that I can test the backend endpoints

## Improvements
<!-- Format: - [ ] Brief description | Context/requirements -->
- [x] Power curve table in power-tab.tsx, should only display data for 5s, 15s, 30s, 1m, 5m, 10m, 20m, 30m, 45m, 1h, and including 2 more columns: heart rate, elevation
- [x] Key Power Outputs: should display 2 rows, 1st for power data, second for IS, TSS and VI
- [x] Create a primary key on activity data: main fields should not be the same. propose a combination. | `UNIQUE(user_id, start_time, sport, distance, duration)`

## Bugs
<!-- Format: - [ ] Brief description | Details/reproduction steps -->
- [x] in UI, clicking on tabs other than overview does not display new data. overview data seems persistant
- [x] when opening the "laps" tab {laps-tab.tsx}, im getting "Failed to execute 'json' on 'Response': Unexpected end of JSON input"

## Tech Debt
<!-- Format: - [ ] Brief description | Why it matters -->
- [x] Create a script to wipe the database clean of any uploaded or generated data. Static data like training zone and HR zones should remain. | `make reset-activities`
- [x] Remove console.log from api.ts | Added for debugging upload flow

## Ideas / Future
<!-- Format: - [ ] Brief description | Notes -->

