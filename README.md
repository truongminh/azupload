# azupload
A simple HTTP gateway to Azure Blob Storage


# Environment
```env
HTTP_PORT=3000
AZURE_BLOB_ACCOUNT_NAME=
AZURE_BLOB_ACCOUNT_KEY=
AZURE_BLOB_CONTAINER=
AZURE_BLOB_PREFIX=azupload
BASE_URL=/files
```
# Usage
- Upload a file: post form `file` to `/<path>` 
If the request succeed, the server replies with a url of the form `/azupload/<path>/<uuid>/<filename>`.
The file can be get using the url.
