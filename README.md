# PKID

PKID is a public Key Indexed Datastore. You can save plain or encrypted data in a public key index; as long as you are the owner of the secret corresponding to that public key.

## Routes

### Set document

```api
POST /{pk}/{project}/{key}
```

Set the value of a document corresponding to {key} inside a {project} indexed by the public key {pk}. This is only possible when sending following header; signed by the private key corresponding to {pk}.

pk is hex encoded;
request data is a base64 encoded and signed;

```json
{ "is_encrypted": true, "payload": "document value", "data_version": 1}
```

header is base64 encoded and signed;

```json
{ "intent": "pkid.store", "timestamp": "epochtime"}
```

### Get document

```api
GET /{pk}/{project}/{key}
```

Get the value of a document corresponding to {key} inside a {project} indexed by the public key {pk}. There is no requirement for a security header

pk is hex encoded;
response data is base64 encoded;

### Delete document

```api
DELETE /{pk}/{project}/{key}
```

Delete the value of a document corresponding to {key} inside a {project} indexed by the public key {pk}. There is no requirement for a security header

pk is hex encoded;

### List

```api
GET /{pk}/{project}
```

Get the keys of a {project} indexed by the public key {pk}. There is no requirement for a security header

pk is hex encoded;
response data is base64 encoded;

## How to run locally

- Run the server on port 3000 with DB file path "pkid.db"

```bash
task start
```

Or manually:

```bash
go run main.go -f pkid.db -p 3000
```

## Test

- Run the app

```bash
task test
```

## Coverage

- Run the app

```bash
task coverage
```

## GO PKID client

- This is a go client for pkid to be able to use pkid

### How to use

```go
import "github.com/rawdaGastan/pkid/client"

privateKey, publicKey := GenerateKeyPair()
serverUrl := "http://localhost:3000"
timeout := 5 * time.Second
pkidClient := NewPkidClient(privateKey, publicKey, serverUrl, timeout)

err := pkidClient.Set("pkid", "key", "value", true)
value, err := pkidClient.Get("pkid", "key")
keys, err := pkidClient.List("pkid")
err = pkidClient.Delete("pkid", "key")
```

### Using PKID in combination with the Threefold Connect app - derived seed scope

- Get the derived seed from TF login
- Generate the key pair using the derived seed

```go
import "github.com/rawdaGastan/pkid/client"

seed := <your seed>
privateKey, publicKey, err := GenerateKeyPairUsingSeed(seed)
serverUrl := "http://localhost:3000"
timeout := 5 * time.Second
pkidClient := NewPkidClient(privateKey, publicKey, serverUrl, timeout)
```
