@ECHO Starting dev server...
REM Edit this value
SET url=staging.my_app.appspot.com
dev_appserver.py --default_gcs_bucket_name %url% dev.yaml