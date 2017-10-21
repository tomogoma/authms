# Deployment

## Pre-requisites

1. Setup and authenticated [gcloud tools](https://cloud.google.com/sdk/docs/) to
deploy to your appengine project.
1. A cloud postgresql instance:
[quick start](https://cloud.google.com/sql/docs/postgres/quickstart)

## Notes

1. All configurations refered to in this readme will be found in [cmd/gcloud]()
1. Deployment should also be done from the [cmd/gcloud]() folder.

```bash
cd cmd/gcloud
```

# Configuration and deployment

## Configuring the service

1. Edit the configuration file at `conf/<name><version>.conf.yml`.
It is documented for required values.
2. Edit [app.yaml](app.yaml) It is documented for required values.

## Deployment

```bash
gcloud app deploy
```

The API docs for service access can be accessed on your browser via
```
https://PROJECT_ID/<version>/<name>/docs
```
Replace PROJECT_ID with your AppEngine project-id.
    