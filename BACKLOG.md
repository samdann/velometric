# Velometric Backlog

## Bugs
<!-- Format: - [ ] Brief description | Details/reproduction steps -->
- [x] in UI, clicking on tabs other than overview does not display new data. overview data seems persistant


## Improvements
<!-- Format: - [ ] Brief description | Context/requirements -->
- [ ] Create a primary key on uploaded ride so that duplicates are not possible (main field should not be the same). propose a combination


## Tech Debt
<!-- Format: - [ ] Brief description | Why it matters -->
- [x] Create a script to wipe the database clean of any uploaded or generated data. Static data like training zone and HR zones should remain. | `make reset-activities`
- [x] Remove console.log from api.ts | Added for debugging upload flow
<!-- [ ] Add proper error handling for API calls | Currently minimal error messages -->



## Ideas / Future
<!-- Format: - [ ] Brief description | Notes -->

