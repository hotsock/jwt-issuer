# JSON Web Token (JWT) Issuer

This is a [Serverless Application Model (SAM)](https://aws.amazon.com/serverless/sam/) application that provides a Lambda function for signing and issuing JSON Web Tokens (JWTs) with an asymmetric key stored in **either** [AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html) or [AWS Key Management Service (KMS)](https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html#asymmetric-cmks) using the ECDSA_SHA_256 (ES256) signing algorithm.

It's designed for easy use with [Hotsock](https://github.com/hotsock/hotsock), but can securely issue JWTs for anything.

This service does not provide functionality for token verification. Instead, the public key is provided in the stack output, which can be used to verify tokens by any external service.

There are two configuration modes for key custody: Parameter Store and KMS. The API is identical for both modes, so there are no application-level design considerations for mode selection.

### Parameter Store (default)

With this mode, your private key material is generated during installation and stored in [Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html) as a `SecureString`. Its value is encrypted with KMS using the default AWS managed key.

When the JWT Issuer Lambda function (cold) starts, it loads the private key value into memory from Parameter Store and uses it to sign keys for the lifetime of that Lambda execution environment.

You'll grant your internal applications access to invoke this Lambda function and receive signed JWTs _without_ granting them access to the stored private key.

This mode is fast, cost-effective, and secure enough for most cases. Why "secure enough"? It's possible that the private key value could be leaked, modified, or deleted. Whether a bug in code, a bad actor in your AWS account, or a permissions mis-configuration, there are no service-level guarantees on the privacy and integrity of the stored key.

#### Performance

Each JWT signing operation requires a call to invoke Lambda. Each Lambda function invocation to sign a JWT takes less than 2ms. Cold-start invocations take about 175ms.

#### Cost

There are no baseline costs when standing up a stack in Parameter Store mode. Everying is usage-based.

Monthly Cost assuming 1,000,000 signed tokens (us-west-2 pricing example):

- $0.20: Lambda requests ($0.20 per 1M requests)
- $0.0034: Lambda duration ($0.0000000017 per 1ms)
- KMS (decrypt) is called once for each Lambda cold start ($0.03 per 10,000 KMS requests). Actual cold start count is very workload dependent so your mileage may vary, but for a real-world instance of this function serving 400 million invocations per month, the KMS decrypt bill is less than $10 per month.

### KMS

KMS mode provides additional security. The private key material never leaves the KMS service in your AWS account, ensuring only AWS principals explicitly authorized with `kms:Sign` permissions for this key can ever generate digital signatures with this key. Even with this permission granted, no one can ever access the underlying private key. A KMS customer managed key (CMK) is created during stack installation and is used for all signing requests.

Since each JWT must be signed and the private key is not directly accessible, each Lambda invocation must call KMS. This adds some runtime latency for each signing operation and KMS calls incur additional costs.

KMS key material can never be modified and if a key is deleted, there is a deletion recovery period to ensure accidental deletion is not permanent. If your company or organization has key compliance requirements, this is probably the best option for you.

#### Performance

Each JWT signing operation requires a call to Lambda, which calls KMS to generate a token signature. Each function invocation takes ~15ms in Lambda. Cold-start invocations take about 200ms. KMS has a default quota of 300 requests per second for ECC signing operations, so be sure to request an increase if you need more than that.

#### Cost

Standing up a stack in your AWS account creates a KMS key, which incurs a charge for its ongoing management. Other than the key management, everything is usage-based.

Monthly Cost assuming 1,000,000 signed tokens (us-west-2 pricing example):

- $1.00: KMS key management
- $15.00: 1,000,000 KMS asymmetric signing requests ($0.15 per 10,000 requests)
- $0.20: Lambda requests ($0.20 per 1M requests)
- $0.13: Lambda duration ($0.0000000067 per 1ms)

As you can see, most of the cost is in KMS. If you're signing a billion tokens each month, this might become cost prohibitive.

## Installation

Launch a stack in your AWS account in less than 5 minutes. Installs using CloudFormation to any of the following regions.

The only option you need to consider is the `KeyCustodianParameter`. Choose `ParameterStore` or `KMS` based on your assessment above, compliance requirements, etc. Other than that, CloudFormation defaults should be fine as you step through the stack creation process.

| Region                    | Alias          | Launch URL                                                                                                                                                                                                                                 |
| ------------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| US East (N. Virginia)     | us-east-1      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-us-east-1.s3.us-east-1.amazonaws.com/jwt-issuer-v1.x.yml)                |
| US East (Ohio)            | us-east-2      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=us-east-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-us-east-2.s3.us-east-2.amazonaws.com/jwt-issuer-v1.x.yml)                |
| US West (N. California)   | us-west-1      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-us-west-1.s3.us-west-1.amazonaws.com/jwt-issuer-v1.x.yml)                |
| US West (Oregon)          | us-west-2      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-us-west-2.s3.us-west-2.amazonaws.com/jwt-issuer-v1.x.yml)                |
| Africa (Cape Town)        | af-south-1     | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=af-south-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-af-south-1.s3.af-south-1.amazonaws.com/jwt-issuer-v1.x.yml)             |
| Asia Pacific (Hong Kong)  | ap-east-1      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-east-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-east-1.s3.ap-east-1.amazonaws.com/jwt-issuer-v1.x.yml)                |
| Asia Pacific (Hyderabad)  | ap-south-2     | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-south-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-south-2.s3.ap-south-2.amazonaws.com/jwt-issuer-v1.x.yml)             |
| Asia Pacific (Jakarta)    | ap-southeast-3 | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-3#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-southeast-3.s3.ap-southeast-3.amazonaws.com/jwt-issuer-v1.x.yml) |
| Asia Pacific (Melbourne)  | ap-southeast-4 | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-4#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-southeast-4.s3.ap-southeast-4.amazonaws.com/jwt-issuer-v1.x.yml) |
| Asia Pacific (Mumbai)     | ap-south-1     | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-south-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-south-1.s3.ap-south-1.amazonaws.com/jwt-issuer-v1.x.yml)             |
| Asia Pacific (Osaka)      | ap-northeast-3 | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-3#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-northeast-3.s3.ap-northeast-3.amazonaws.com/jwt-issuer-v1.x.yml) |
| Asia Pacific (Seoul)      | ap-northeast-2 | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-northeast-2.s3.ap-northeast-2.amazonaws.com/jwt-issuer-v1.x.yml) |
| Asia Pacific (Singapore)  | ap-southeast-1 | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-southeast-1.s3.ap-southeast-1.amazonaws.com/jwt-issuer-v1.x.yml) |
| Asia Pacific (Sydney)     | ap-southeast-2 | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-southeast-2.s3.ap-southeast-2.amazonaws.com/jwt-issuer-v1.x.yml) |
| Asia Pacific (Tokyo)      | ap-northeast-1 | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ap-northeast-1.s3.ap-northeast-1.amazonaws.com/jwt-issuer-v1.x.yml) |
| Canada (Central)          | ca-central-1   | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=ca-central-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-ca-central-1.s3.ca-central-1.amazonaws.com/jwt-issuer-v1.x.yml)       |
| Europe (Frankfurt)        | eu-central-1   | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-central-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-central-1.s3.eu-central-1.amazonaws.com/jwt-issuer-v1.x.yml)       |
| Europe (Ireland)          | eu-west-1      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-west-1.s3.eu-west-1.amazonaws.com/jwt-issuer-v1.x.yml)                |
| Europe (London)           | eu-west-2      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-west-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-west-2.s3.eu-west-2.amazonaws.com/jwt-issuer-v1.x.yml)                |
| Europe (Milan)            | eu-south-1     | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-south-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-south-1.s3.eu-south-1.amazonaws.com/jwt-issuer-v1.x.yml)             |
| Europe (Paris)            | eu-west-3      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-west-3#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-west-3.s3.eu-west-3.amazonaws.com/jwt-issuer-v1.x.yml)                |
| Europe (Spain)            | eu-south-2     | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-south-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-south-2.s3.eu-south-2.amazonaws.com/jwt-issuer-v1.x.yml)             |
| Europe (Stockholm)        | eu-north-1     | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-north-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-north-1.s3.eu-north-1.amazonaws.com/jwt-issuer-v1.x.yml)             |
| Europe (Zurich)           | eu-central-2   | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=eu-central-2#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-eu-central-2.s3.eu-central-2.amazonaws.com/jwt-issuer-v1.x.yml)       |
| Israel (Tel Aviv)         | il-central-1   | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=il-central-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-il-central-1.s3.il-central-1.amazonaws.com/jwt-issuer-v1.x.yml)       |
| Middle East (Bahrain)     | me-south-1     | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=me-south-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-me-south-1.s3.me-south-1.amazonaws.com/jwt-issuer-v1.x.yml)             |
| Middle East (UAE)         | me-central-1   | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=me-central-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-me-central-1.s3.me-central-1.amazonaws.com/jwt-issuer-v1.x.yml)       |
| South America (São Paulo) | sa-east-1      | [Launch Stack](https://console.aws.amazon.com/cloudformation/home?region=sa-east-1#/stacks/new?stackName=JWTIssuer&templateURL=https://jwt-issuer-stack-templates-sa-east-1.s3.sa-east-1.amazonaws.com/jwt-issuer-v1.x.yml)                |

AWS GovCloud regions are not currently supported because the regions are missing `provided.al2023` runtime support in Lambda.

The CloudFormation stack will have the status `CREATE_COMPLETE` when the installation is finished. At this point, you can go to the "Outputs" tab in the stack and you'll see the following variables.

### `JWTIssuerFunctionArn`

This is the Amazon Resource Name (Arn) of the Lambda function you'll invoke to sign JWTs. Examples of how to use it in the usage section below. This can be used as the value for `function-name` (CLI) or `function_name` (Ruby SDK) below.

Example: `arn:aws:lambda:us-east-1:111111111111:function:JWTProd-JWTIssuerPSFunction-mUI2JR398C8c`

### `KeyArn`

This is the Amazon Resource Name (Arn) of the KMS key that is used when signing keys. This is left blank if using Parameter Store.

### `KeyID`

When signing tokens, this is the value that the `kid` header claim will be set to in all JWTs. If using Parameter store, it's the UUID in the CloudFormation stack's ARN. If using KMS, it's the UUID in the KMS key ARN.

Example: `ef814598-df45-4aa4-9f32-1b616ae6afda`

### `PublicKeyPEMBase64`

This is the public key in PEM format encoded to Base 64. If you're using [Hotsock](https://github.com/hotsock/hotsock), you can paste this value directly into the `SigningKey1EncodedParameter` or `SigningKey2EncodedParameter` to allow Hotsock to authorize signed keys from this stack.

It's completely harmless for this public key to be passed around. It's named appropriately!

Example: `LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFL2RmYXdYbkZxb0FWTG81NU04UW5yelBpazZOcgpYQnUybllLQkY5YTM2bGZtK0FPcG8xYzhxUzJKQkhYVVV1WE1YajAzdzh0Q1F0bGZidXFaaUljWGVnPT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==`

If you decode this from Base 64 to a string, you'll see it's a PEM-formatted public key.

```
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/dfawXnFqoAVLo55M8QnrzPik6Nr
XBu2nYKBF9a36lfm+AOpo1c8qS2JBHXUUuXMXj03w8tCQtlfbuqZiIcXeg==
-----END PUBLIC KEY-----
```

### `SigningMethod`

This is the JWT signing algorithm. Always set to `ES256`.

### `Version`

The release version of your installation.

Example: `v1.0`

## Usage

First you need to grant your application the ability to invoke the JWT issuer Lambda function using IAM. At a minimum, an IAM policy tied to your application's AWS role or user must have `Allow` set for the `lambda:InvokeFunction` action on the Arn referenced in the `JWTIssuerFunctionArn` output from your installation. If, for example, your application runs on [AWS Fargate](https://aws.amazon.com/fargate/), you'd want to add this permissions policy to the task execution role for your ECS service. If your application runs on [EC2](https://aws.amazon.com/ec2/), you probably need to add this permissions policy to the [IAM role associated with your EC2 instance(s)](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html). You can also use IAM users with hard-coded credentials, but that's not recommended.

Here's a sample policy.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": ["lambda:InvokeFunction"],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:lambda:us-east-1:111111111111:function:JWTProd-JWTIssuerPSFunction-mUI2JR398C8c"
      ]
    }
  ]
}
```

Using the AWS SDK in the language of your choice, call the Lambda invoke API to sign a token. Here's an example using the AWS CLI.

This generates a token with the `aud` and `channels` claims set explicitly and configures the `exp` claim to expire the token 30 seconds after it is issued.

```
aws lambda invoke \
  --function-name JWTIssuer-JWTIssuerPSFunction-MFlF1fyVpWkZ \
  --payload '{"claims":{"aud":"hotsock","channels":{"chat":{"subscribe":true}}},"ttl":30}' \
  --cli-binary-format raw-in-base64-out \
  /dev/stdout
```

The response is JSON and contains the signed token in the `token` field.

```json
{
  "token": "eyJhbGciOiJFUzI1NiIsImtpZCI6ImVmODE0NTk4LWRmNDUtNGFhNC05ZjMyLTFiNjE2YWU2YWZkYSIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJob3Rzb2NrIiwiY2hhbm5lbHMiOnsiY2hhdCI6eyJzdWJzY3JpYmUiOnRydWV9fSwiZXhwIjoxNzEzODM2OTUwfQ.Gz5iLG6O7YBQf8jAJafbaeCUxC08JnVEfnzbPOnn3S90hdiptlztp4Io3UmnhKjTqphf1G1ZYKQ29jbU7C6Xow"
}
```

Here's the same invocation, but using the [Ruby SDK](https://docs.aws.amazon.com/sdk-for-ruby/v3/api/Aws/Lambda/Client.html).

```ruby
Aws::Lambda::Client.new.invoke(
  function_name: "JWTIssuer-JWTIssuerPSFunction-MFlF1fyVpWkZ",
  payload: JSON.dump({"claims":{"aud":"hotsock","channels":{"chat":{"subscribe":true}}},"ttl":30})
).payload.read
# => "{\"token\":\"eyJhbGciOiJFUzI1NiIsImtpZCI6ImVmODE0NTk4LWRmNDUtNGFhNC05ZjMyLTFiNjE2YWU2YWZkYSIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJob3Rzb2NrIiwiY2hhbm5lbHMiOnsiY2hhdCI6eyJzdWJzY3JpYmUiOnRydWV9fSwiZXhwIjoxNzEzODM2OTUwfQ.Gz5iLG6O7YBQf8jAJafbaeCUxC08JnVEfnzbPOnn3S90hdiptlztp4Io3UmnhKjTqphf1G1ZYKQ29jbU7C6Xow\"}"
```

### `claims`

`Object` (required) - Provide all claims here as a JSON object.

### `setIat`

`Boolean` (optional) - If true, sets the `iat` claim to the time that the token was issued. Overrides explicit `iat` set in `claims`. Defaults to `false`.

### `setJti`

`Boolean` (optional) - If true, sets the `jti` claim to a randomly generated UUID (v4). Overrides explicit `jti` set in `claims`. Defaults to `false`.

### `ttl`

`Integer` (optional) - If supplied, sets the token expiration claim (`exp`) to a timestamp this many seconds from when the token is issued. Overrides explicit `exp` set in `claims`. If not supplied, make sure you specify your own `exp` claim in `claims` to ensure the token expires.

## Updates & maintenance

You can assume that v1.x is stable. Updating an existing stack to the latest 1.x may add new functionality, but will not break existing APIs documented in this README, replace AWS resources, or change behavior. The underlying Go code may change at any time, as the code is not intended for use as a library imported into your code.

To update an existing stack, open CloudFormation in the AWS Console.

1. Find your installation's stack (it's called JWTIssuer if you used the default name) and click the "Update" button.
1. On the "Prepare template" screen, choose "Replace current template".
1. For "Template source", use "Amazon S3 URL" and copy the URL for your region from the table below. Click "Next" through the screens that follow keeping all other defaults. Acknowledge any capabilities requirements on the final screen and click "Submit". Stack updates typically take no longer than 2 minutes.

| Region                    | Alias          | Amazon S3 URL URL                                                                                     |
| ------------------------- | -------------- | ----------------------------------------------------------------------------------------------------- |
| US East (N. Virginia)     | us-east-1      | https://jwt-issuer-stack-templates-us-east-1.s3.us-east-1.amazonaws.com/jwt-issuer-v1.x.yml           |
| US East (Ohio)            | us-east-2      | https://jwt-issuer-stack-templates-us-east-2.s3.us-east-2.amazonaws.com/jwt-issuer-v1.x.yml           |
| US West (N. California)   | us-west-1      | https://jwt-issuer-stack-templates-us-west-1.s3.us-west-1.amazonaws.com/jwt-issuer-v1.x.yml           |
| US West (Oregon)          | us-west-2      | https://jwt-issuer-stack-templates-us-west-2.s3.us-west-2.amazonaws.com/jwt-issuer-v1.x.yml           |
| Africa (Cape Town)        | af-south-1     | https://jwt-issuer-stack-templates-af-south-1.s3.af-south-1.amazonaws.com/jwt-issuer-v1.x.yml         |
| Asia Pacific (Hong Kong)  | ap-east-1      | https://jwt-issuer-stack-templates-ap-east-1.s3.ap-east-1.amazonaws.com/jwt-issuer-v1.x.yml           |
| Asia Pacific (Hyderabad)  | ap-south-2     | https://jwt-issuer-stack-templates-ap-south-2.s3.ap-south-2.amazonaws.com/jwt-issuer-v1.x.yml         |
| Asia Pacific (Jakarta)    | ap-southeast-3 | https://jwt-issuer-stack-templates-ap-southeast-3.s3.ap-southeast-3.amazonaws.com/jwt-issuer-v1.x.yml |
| Asia Pacific (Melbourne)  | ap-southeast-4 | https://jwt-issuer-stack-templates-ap-southeast-4.s3.ap-southeast-4.amazonaws.com/jwt-issuer-v1.x.yml |
| Asia Pacific (Mumbai)     | ap-south-1     | https://jwt-issuer-stack-templates-ap-south-1.s3.ap-south-1.amazonaws.com/jwt-issuer-v1.x.yml         |
| Asia Pacific (Osaka)      | ap-northeast-3 | https://jwt-issuer-stack-templates-ap-northeast-3.s3.ap-northeast-3.amazonaws.com/jwt-issuer-v1.x.yml |
| Asia Pacific (Seoul)      | ap-northeast-2 | https://jwt-issuer-stack-templates-ap-northeast-2.s3.ap-northeast-2.amazonaws.com/jwt-issuer-v1.x.yml |
| Asia Pacific (Singapore)  | ap-southeast-1 | https://jwt-issuer-stack-templates-ap-southeast-1.s3.ap-southeast-1.amazonaws.com/jwt-issuer-v1.x.yml |
| Asia Pacific (Sydney)     | ap-southeast-2 | https://jwt-issuer-stack-templates-ap-southeast-2.s3.ap-southeast-2.amazonaws.com/jwt-issuer-v1.x.yml |
| Asia Pacific (Tokyo)      | ap-northeast-1 | https://jwt-issuer-stack-templates-ap-northeast-1.s3.ap-northeast-1.amazonaws.com/jwt-issuer-v1.x.yml |
| Canada (Central)          | ca-central-1   | https://jwt-issuer-stack-templates-ca-central-1.s3.ca-central-1.amazonaws.com/jwt-issuer-v1.x.yml     |
| Europe (Frankfurt)        | eu-central-1   | https://jwt-issuer-stack-templates-eu-central-1.s3.eu-central-1.amazonaws.com/jwt-issuer-v1.x.yml     |
| Europe (Ireland)          | eu-west-1      | https://jwt-issuer-stack-templates-eu-west-1.s3.eu-west-1.amazonaws.com/jwt-issuer-v1.x.yml           |
| Europe (London)           | eu-west-2      | https://jwt-issuer-stack-templates-eu-west-2.s3.eu-west-2.amazonaws.com/jwt-issuer-v1.x.yml           |
| Europe (Milan)            | eu-south-1     | https://jwt-issuer-stack-templates-eu-south-1.s3.eu-south-1.amazonaws.com/jwt-issuer-v1.x.yml         |
| Europe (Paris)            | eu-west-3      | https://jwt-issuer-stack-templates-eu-west-3.s3.eu-west-3.amazonaws.com/jwt-issuer-v1.x.yml           |
| Europe (Spain)            | eu-south-2     | https://jwt-issuer-stack-templates-eu-south-2.s3.eu-south-2.amazonaws.com/jwt-issuer-v1.x.yml         |
| Europe (Stockholm)        | eu-north-1     | https://jwt-issuer-stack-templates-eu-north-1.s3.eu-north-1.amazonaws.com/jwt-issuer-v1.x.yml         |
| Europe (Zurich)           | eu-central-2   | https://jwt-issuer-stack-templates-eu-central-2.s3.eu-central-2.amazonaws.com/jwt-issuer-v1.x.yml     |
| Israel (Tel Aviv)         | il-central-1   | https://jwt-issuer-stack-templates-il-central-1.s3.il-central-1.amazonaws.com/jwt-issuer-v1.x.yml     |
| Middle East (Bahrain)     | me-south-1     | https://jwt-issuer-stack-templates-me-south-1.s3.me-south-1.amazonaws.com/jwt-issuer-v1.x.yml         |
| Middle East (UAE)         | me-central-1   | https://jwt-issuer-stack-templates-me-central-1.s3.me-central-1.amazonaws.com/jwt-issuer-v1.x.yml     |
| South America (São Paulo) | sa-east-1      | https://jwt-issuer-stack-templates-sa-east-1.s3.sa-east-1.amazonaws.com/jwt-issuer-v1.x.yml           |

Note: The above URLs will appear to not work if clicked on from a browser. They are only meant for use within CloudFormation. These templates are generated and written to S3 in all regions from GitHub Actions ([.github/workflows/regional_templates.yml](.github/workflows/regional_templates.yml)) when new releases are tagged.

### Switch from Parameter Store to KMS or vice versa

**Switching key custodians is not recommended.** Technically, switching the `KeyCustodianParameter` and updating the stack will do the right thing and change your preference. If you switch this way, your private/public keys will be deleted from KMS/Parameter Store during the update, the Lambda function used to sign keys will be replaced (and will have a different Arn in `JWTIssuerFunctionArn`), and anything still attempting to sign with the previous keys will stop working immediately.

Instead, the recommendation is to launch a new stack that uses the desired service for key custody (Parameter Store or KMS). You can begin signing keys with the new installation immediately and delete the old stack once you've verified it is no longer needed.

## Local development & manual builds

To develop and test locally or to deploy a manual build, clone this repository and install the following.

- Install [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- Install [Go](https://go.dev/doc/install) (the latest release should work)

Run tests with `make test`. Build all binaries for deployment on Lambda with `make build`.

Use `sam deploy --guided` to package local CloudFormation and deploy to a new stack using your AWS CLI credentials.
