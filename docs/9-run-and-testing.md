#  Chapter 9: Running and Testing the Service

In this chapter, we will go over how to run your Guild Service and perform some basic tests to 
ensure that everything is working as expected.

# 9.1 Running the Service

## Setup

To be able to run this sample app, you will need to follow these setup steps.

- Create a docker compose `.env` file by copying the content of [.env.template](.env.template) file.
- Fill in the required environment variables in `.env` file as shown below.

   ```txt
   AB_BASE_URL=https://demo.accelbyte.io      # Base URL of AccelByte Gaming Services demo environment
   AB_CLIENT_ID='xxxxxxxxxx'                  # Use Client ID from the Setup section
   AB_CLIENT_SECRET='xxxxxxxxxx'              # Use Client Secret from the Setup section
   AB_NAMESPACE='xxxxxxxxxx'                  # Use Namespace ID from the Setup section
   PLUGIN_GRPC_SERVER_AUTH_ENABLED=false      # Enable or disable access token and permission verification
   ```

   > :warning: **Keep PLUGIN_GRPC_SERVER_AUTH_ENABLED=false for now**: It is currently not
   supported by AccelByte Gaming Services but it will be enabled later on to improve security. If it is
   enabled, the gRPC server will reject any calls from gRPC clients without proper authorization
   metadata.

- Ensure `grpc-gateway-dependencies` mentioned in [chapter 4](4-installation-and-setup.md) is up and running

## Building

To build this sample app, use the following command.

```
make build
```

To build and create a docker image of this sample app, use the following command.

```
make image
```

For more details about these commands, see [Makefile](Makefile).

## Running

To run the existing docker image of this sample app which has been built before, use the following command.

```
docker-compose up
```

OR

To build, create a docker image, and run this sample app in one go, use the following command.

```
docker-compose up --build
```

## Testing

After starting the service, you can test it to make sure it's working correctly.

We will use curl command to test our service. For example, to test `CreateOrUpdateGuildProgress` endpoint, you can run:

```bash
$ curl -X POST http://localhost:8000/guild/v1/progress \
    -H 'Content-Type: application/json' \
    -d '{
      "guild_id": "my-guild-id",
      "guild_progress": {
        "objectives": {
          "quest1": 100,
          "quest2": 200
        }
      }
    }'
```

And to test `GetGuildProgress` endpoint:

```bash
$ curl -X GET http://localhost:8000/guild/v1/progress/my-guild-id
```

You should see the updated guild progress in the response.

