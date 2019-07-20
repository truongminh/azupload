# azupload
A simple HTTP gateway to Azure Blob Storage


# Environment
HTTP_PORT=3000
AZURE_BLOB_ACCOUNT_NAME=
AZURE_BLOB_ACCOUNT_KEY=
AZURE_BLOB_CONTAINER=
AZURE_BLOB_PREFIX=azupload
BASE_URL=/files

# Usage
- Upload a file: post form `file` to `/` 
If the request succeed, a url is created.
Just get the content using the new url

