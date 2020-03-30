# XYZ Service API

This service represents the XYZ API gateway for all XYZ API endpoints.

Run the application using:

```
FIREBASE_SERVICE_ACCOUNT_KEY_FILE=./serviceAccountKey.json make run
```

or
```
FIREBASE_SERVICE_ACCOUNT_KEY_FILE=/home/luigi/base/xyz/serviceAccountKey.json ./xyz
```

Run the API tests in a second shell:

```
make test
```

## Documentation

All API endpoints are documented using Swagger docs.

The Swagger API doc file is in `.docs/`
