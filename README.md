# tail-pubsub
Tail a Google PUB/SUB topic

## Usage:
```
$ ./tail-pubsub -h
Usage of ./tail-pubsub:
  -batchsize=1: How many messages to get at once
  -project="": Google project-name
  -topic="": PUB/SUB topic name to subscribe to
```

## Credentials:
Credentials can either be derived from a credentials file, or from instance permissions:
* Credentials-file should be '~/.config/gcloud/application_default_credentials.json'
* Alternatively you can use any other file:
  ```export GOOGLE_APPLICATION_CREDENTIALS=~/Downloads/credentials-file.json```
* Instance permissions required:
  * PubsubScope = "https://www.googleapis.com/auth/pubsub"
