# AppEngine Go Template

Boilerplate template for my typical AppEngine apps.

## Included Functionality

* Basic account setup (most likely needs to be tweaked per app)
* Authentication
* CORS request handline
* Google Cloud Storage uploading
* Google Cloud Storage image lazy-resizing

## Getting Started

* Clone the repo and delete the `.git` folder.
* Set the `application` name to app's name within the `app.yaml` file
* Set the `ALLOWED_ORIGINS` value in the app.yaml file. If not using CORS, make it blank.