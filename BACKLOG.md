# Velometric Backlog

## Bugs
<!-- Format: - [ ] Brief description | Details/reproduction steps -->
- [x] in UI, clicking on tabs other than overview does not display new data. overview data seems persistant


## Improvements
<!-- Format: - [ ] Brief description | Context/requirements -->
- [ ] Power curve table in power-tab.tsx, should only display data for 5s, 15s, 30s, 1m, 5m, 10m, 20m, 30m, 45m, 1h, and including 2 more columns: heart rate, elevation
- [ ] Key Power Outputs: should display 2 rows, 1st for power data, second for IS, TSS and VI
- [ ] Create a primary key on activity data: main fields should not be the same. propose a combination.


## Tech Debt
<!-- Format: - [ ] Brief description | Why it matters -->
- [x] Create a script to wipe the database clean of any uploaded or generated data. Static data like training zone and HR zones should remain. | `make reset-activities`
- [x] Remove console.log from api.ts | Added for debugging upload flow
<!-- [ ] Add proper error handling for API calls | Currently minimal error messages -->



## Ideas / Future
<!-- Format: - [ ] Brief description | Notes -->

