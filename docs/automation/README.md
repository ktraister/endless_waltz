# Automation
This page is intended to inform the reader of automation written for
the EW Circut and messenger.

# Platform
GitHub Actions serves as the automation server for EW. All automations
live within `endless_waltz/.github/workflows`. 

## CI/CD
Full CI/CD is in place for `exchange`, `random`, `webapp`. Build and 
deploy of the container is handled in the respective workflow. 

## CI
CI ONLY is handled for `reaper` and `ew-entropy`. Container is built, 
but not deployed due to the sensitive nature of the reaper service.

## Merge Checks
Merge checks are performed on application code before allowing merge. 
These checks are in place for all services. 

## Database Backups
Nightly backups are performed and uploaded to s3 from Github Actions.
