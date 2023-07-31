# energy-smart-booking

This repo is a collection of services and shared code related to the smart
meter installation booking journey.

## Repo structure
**cmd/** - all individual commands, implementations, and business logic.

**internal/** - common utilities like database connection code, shared domain, types, repositories, shared test utilities, etc

**Makefile** - single makefile taking in source_files and services as arguments to build a specific executable
    example: make SERVICE=energy-smart-booking-opt-out SOURCE_FILES=cmd/opt-out/ energy-smart-booking-opt-out

## GitHub workflows

1. Reusable build service [build_service_reusable_workflow](https://github.com/utilitywarehouse/energy-smart-booking/blob/master/.github/workflows/build_service_reusable_workflow.yml)
    Run install, lint, test and docker build for a given service.
    This workflow is intended to run only as a callable action by a specific booking service workflow or to build all in tandem.

2. Callable build-smart-booking workflow [smart_booking_workflow](https://github.com/utilitywarehouse/energy-smart-booking/blob/master/.github/workflows/energy_smart_booking.yml)
    Calls reusable build service for all booking service executables.
    Triggered when there is any change in the project other than just cmd/$BOOKING_SERVICE.

3. CI entrypoint (https://github.com/utilitywarehouse/energy-smart-booking/blob/master/.github/workflows/ci.yml).
    Checks the files changed compared to previous commit to determine what build it needs to run.
    Logic:
    - use tj-actions/changed-files@v35 to get all changed files
    - go through each file and:
      - if file belongs to cmd/$BOOKING_SERVICE_NAME - set job output as needs.check.outputs.'$BOOKING_SERVICE_NAME' = true
      - if file doesn't belong to any cmd/** - set job output as needs.check.outputs.booking_services = true

    - job entry for each specific fabricator checking if it should run using:

    `needs.check.outputs.booking_services != 'true' && needs.check.outputs.opt_out == 'true'`

    - job entry to build all booking checking if it should run using:

    `needs.check.outputs.booking_services == 'true'`

    By using **workflow_dispatch** we avoid duplication of workflows triggered when there is a PR opened and commits are done on the feature implementation.
   - every branch is built on push without a need to create a PR
   - when you create PR this “branch build” is already picked up by checks
   - when you update PR it is built, because branch has changed
   - master is built on branch merge after PR

## Booking Services Overview

## Preparing the Repository
Because we are using gomock, some of the test files import the generated mocks, however they are not part of the tracked changes(to avoid us dealing with generated code susceptible to change and we do not want to be concerned with merge conflicts).
In order to be able to run commands such as `go mod` or even `go test` you should run `go generate` to create all the necesary mocks.

### opt-out

TBC

