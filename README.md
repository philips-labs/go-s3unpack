# go-s3unpack
Unpacks ZIP files on an S3 bucket

# Usage
Exposes an endpoint `/unpack` which accepts POST requests of the following JSON:

```json
{
  "sourceFile": "zipfolder/Manhattan.zip",
  "destinationPath": "unpacked"
}
```

# curl example

```shell
> curl -H "Content-Type: application/json" -X POST http://localhost:8080/unpack -d '{"sourceFile":"zipfolder/Manhattan.zip", "destinationPath":"unpacked"}'
```

# Deployment
This app is compatible with HSDP Cloud foundry and expects an S3 bucket to be bound. The app will connec to this bucket and listen for `POST /unpack` requests.

