# Velometric Backlog

## Bugs
<!-- Format: - [ ] Brief description | Details/reproduction steps -->


## Improvements
<!-- Format: - [ ] Brief description | Context/requirements -->
- [ ] Support FIT files from multiple sources | Garmin Connect (ZIP), Strava, Wahoo may have different formats/fields


## Tech Debt
<!-- Format: - [ ] Brief description | Why it matters -->
- [x] Create a script to wipe the database clean of any uploaded or generated data. Static data like training zone and HR zones should remain. | `make reset-activities`
- [x] Remove console.log from api.ts | Added for debugging upload flow
<!-- [ ] Add proper error handling for API calls | Currently minimal error messages -->



## Ideas / Future
<!-- Format: - [ ] Brief description | Notes -->

