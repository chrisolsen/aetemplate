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
* Set the `ALLOWED_ORIGINS` value in the dev.yaml and app.yaml file. If not using CORS, make it blank.
* Create Google Cloud Storage default app buckets and update the dev.bat file bucket name

## Appengine SSL Certs

1. Open bash shell run `certbot-auto certonly -a manual -d my-domain.com -d www.my-domain.com`
2. Follow the steps to verify domain (see the app.yml and app.go file for existing handlers used to verify ownership)
3. Navgate to `/etc/letsencrypt/live/rallyup.io`. (You will need to `sudo su` to get into `live/..`)
4. [Update certs](https://console.cloud.google.com/appengine/settings/certificates?project=[project-id]&serviceId=default) with the new `cert.pem` `privkey_rsa.pem`.
    * Note: The `privkey.pem` must be converted to rsa by `openssl rsa -in privkey.pem -out privkey_rsa.pem` for Appengine to accept it as valid.
5. Set the full_chain.pem file contents into `PEM encoded X.509 public key certificate` and the above rsa encoded `.pem` content into the `Unencrypted PEM encoded RSA private key`