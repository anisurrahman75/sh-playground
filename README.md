# sh-playground

### overview

This project allows you to interact with both Amazon S3 and Google Cloud Storage (GCS). Before running the code, you'll need to set up your credentials for both cloud providers.

**Setup Credentials**

```bash
# S3:
$ export AWS_ACCESS_KEY_ID=<your_access_key_id>
$ export AWS_SECRET_ACCESS_KEY=<your_secret_access_key>

# GCS:
$ export GOOGLE_PROJECT_ID="appscode-testing"
$ export GOOGLE_APPLICATION_CREDENTIALS=<path_to_your_credentials_file>
```

**Running the Application**

Once the credentials are set, you can execute the Go application using:
```bash
$ go run .
```