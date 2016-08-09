mackerel-plugin-aws-billing
=================================
AWS billing custom metrics plugin for mackerel.io agent.

## Synopsis

```shell
mackerel-plugin-aws-aws-billing [-region=<aws-region>] [-access-key-id=<id>] [-secret-access-key=<key>] [-tempfile=<tempfile>]
```
* if you run on an ec2-instance and the instance is associated with an appropriate IAM Role, you probably don't have to specify `-access-key-id` & `-secret-access-key`

## AWS IAM Policy
the credential provided manually or fetched automatically by IAM Role should have the policy that includes an action, 'cloudwatch:GetMetricStatistics'

## Enable Billing Alert
mackerel-plugin-aws-billing needs to enable billing alerts. So, turn on 'Receive Billing alert' on AWS console.

## Example of mackerel-agent.conf

```
[plugin.metrics.aws-billing]
command = "/path/to/mackerel-plugin-aws-billing"
```
