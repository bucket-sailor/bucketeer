# Bucketeer

![demo](./assets/demo.gif)
*Browsing a [Cloudflare R2](https://www.cloudflare.com/developer-platform/r2/) bucket containg the Linux kernel source tree.*

## Getting Started

Assuming you have the [AWS CLI](https://aws.amazon.com/cli/) installed and configured.

```shell
bucketeer my-bucket
```

If you aren't using AWS you will need to specify the S3 endpoint to use (funnily enough AWS doesn't provide a standard configuration option for this):

```shell
bucketeer --endpoint-url=https://my-account.r2.cloudflarestorage.com my-bucket
```

## Features

* Easy to use Web UI.
* Run locally or as a container (headless).
* Upload/download files (without limits).
* Large directory support.

## Telemetry

By default Bucketeer gathers anonymous crash and usage data. This anonymized data is processed on our servers within the EU and is not shared with third parties. You can opt out of telemetry by setting the `BUCKETEER_NO_TELEMETRY=1` environment variable.

## License

Bucketeer is dual licensed under the [AGPLv3](./LICENSE) and a commercial license.

The commercial distribution includes additional features such as authentication/authorization, audit logging, and includes an accompanying Kubernetes operator to make deploying Bucketeer a breeze.

If you'd like to use the commercial distribution please reach out to me via [Email](mailto:damian@pecke.tt). Don't worry, its not going to cost you an arm and a leg. All sales go a long way to helping keep Bucketeer awesome.