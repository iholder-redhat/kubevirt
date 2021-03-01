# How to update Go version 
A quick guide to update Kubevirt's Go version.

To update the Go version we need to update the builder image so that it uses the new version,
push it to the registry and finally let Kubevirt use the new builder image.

In addition, Bazel has to be updated if the current version does not support the target Go version.

## Updating Go Version
### Updating builder image
* Update `VERSION` tag in [hack/builder/version.sh](../hack/builder/version.sh).
  * For example, change from `VERSION=30-9.0.3` to `VERSION=30-9.0.4`
* Change the `GIMME_GO` version in the [hack/builder/Dockerfile](../hack/builder/Dockerfile) to the desired Go version.
* Rebuild the builder image by executing `make builder-build`.
  
### Publishing builder image
* Publish new builder image with `make builder-publish`.
  * Note: Proper access rights are required in order to publish builder image.
  * When publish is finished, a SHA ID will be presented. It should look similar to:
    ```shell
    ...
    + docker manifest create --amend quay.io/kubevirt/builder:30-9.0.4 quay.io/kubevirt/builder:30-9.0.4-amd64
    Created manifest list quay.io/kubevirt/builder:30-9.0.4
    + docker manifest push quay.io/kubevirt/builder:30-9.0.4
    sha256:0e6807882debf6e3b482628cef2ef79493f99201d04e6000156545a2233117f2
    + cleanup
    + rm manifests/ -rf
    ```
* Change `KUBEVIRT_BUILDER_IMAGE` variable in [hack/dockerized](../hack/dockerized) to the SHA ID from the previous step.
* In [WORKSPACE](../WORKSPACE) change `go_version` to the desired Go version.
  * Should look similar to:
    ```shell
    go_register_toolchains(
        go_version = "1.14.14",
        nogo = "@//:nogo_vet",
    )
    ```

## Update Bazel Version
* In [WORKSPACE](../WORKSPACE) find current Bazel release's SHA ID which can be found under `io_bazel_rules_go`.
  * Should look similar to:
    ```shell
    http_archive(
      name = "io_bazel_rules_go",
      sha256 = "52d0a57ea12139d727883c2fef03597970b89f2cc2a05722c42d1d7d41ec065b",
      urls = [
        ...
      ],
    )
    ```
* Visit [Bazel's releases page](https://github.com/bazelbuild/rules_go/releases) and check whether the current Bazel version supports the new Go version.
  * If it is not supported, replace the `io_bazel_rules_go` definition with the one provided in Bazel's page.
* Use [project-infra's uploader tool](https://github.com/kubevirt/project-infra/blob/master/plugins/cmd/uploader/README.md) to upload new dependencies to dependency mirror.


  
