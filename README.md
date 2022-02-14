# Lambda Invocation Stats

Lists the first 50 lambdas per AWS region with an invocation count for the last
28 days.

## Usage
```
Usage:
  lambda-invocations [flags]

Flags:
  -d, --debug            Debug
  -e, --env              Use environment vars for credentials
  -h, --help             help for lambda-invocations
  -p, --profile string   AWS credentials profile name (default "default")
  -r, --region string    AWS region (e.g. us-east-1). Use "all" for all regions (default "eu-west-2")
```

### Profile Flag -p
If you have multiple AWS profiles, use the -p flag to specify which one to use.
Defaults to "default" profile

### Region Flag -r
Specify a region or use "all" to enumerate all regions.
Defaults to eu-west-2.

### Debug Flag -d
Outputs debugging information

### MFA Accounts
If your account is secured with MFA, you may need to generate a short lived
access token and export environment variables. If this is the case, use the
-e flag to use the environment variables.

Example of how to generate access token:
```
aws sts get-session-token --serial-number <arn-for-mfa> --duration-seconds 3600 --token-code <token-from-device> --profile <if-not-default-profle>
```



## Example output
```
3 Lambda functions found in eu-west-2
+----------------+-----------+-------------+
| FUNCTION       | TYPE      | INVOCATIONS |
+----------------+-----------+-------------+
| start-instance | python3.9 |          23 |
| lambda-fun     | go1.x     |       34936 |
| test-go        | go1.x     |       17737 |
+----------------+-----------+-------------+
```

## Errors
```
Could not get lambdas in <region>
```
This is usually caused because your account does not have access to that region.
