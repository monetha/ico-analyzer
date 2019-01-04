# ICO Analyzer

## Prerequisites for testing Lambda function locally

* Docker
* AWS SAM CLI

The account that you are running the commands with should be able to manager Docker (e.g., be member of the `docker` group or use `sudo`)

## Install AWS SAM CLI

To install AWS SAM CLI, run the following command:

```shell
pip install --upgrade aws-sam-cli
```

For more details please go to https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html

## Build locally

To build the code locally please use the following command:

```shell
CI_PIPELINE_ID=local build/linux-build.sh
```

## Test locally

Please build the code as explained in the previous section. After the code is built, run the following command to start the local API Gateway instance:

```shell
sam local start-api
```

After the API Gateway container is up and running you, can call the Lambda function using `http://127.0.0.1:3000/`

For more details on testing and debugging the Lambda function locally, please check https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-using-debugging-golang.html

## Examples


When the API Gateway is up and running, you can use the following commands to check if its working:

```shell
curl --request GET \
  --url http://127.0.0.1:3000/ 
```

```shell
curl --request POST \
  --url http://127.0.0.1:3000/ \
  --header 'Content-Type: application/json' \
  --data '{"name": "YourName"}'
```