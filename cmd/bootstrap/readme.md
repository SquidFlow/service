# Bootstrap CLI

Bootstrap is a command-line tool for initializing H4 Bootstrap.

## Quick Start

1. Set up your Kubernetes configuration:

   ```shell
   $ bootstrap -c deploy/service/templates/config.toml config set kubernetes.mode ~/.kube/config # show same config file with backend service
   ```

2. Configure your application repository:

   ```shell
   $ bootstrap -c deploy/service/templates/config.toml config set application_repo git@github.com:h4-poc/application_repo.git
   ```

3. Initialize the bootstrap environment:

   ```shell
   $ bootstrap -c deploy/service/templates/config.toml init
   ```

## Available Commands

- `init`: Initialize the bootstrap environment
- `project`: Project mangem
- `status`: Check the current status of the bootstrap environment

## Detailed Usage

### Configuration

Set up your environment using the `config set` command:

   ```shell
   $ export BOOTSTRAP_CONFIG=deploy/service/templates/config.toml # option, if not set this ENV, should pass -c <config-path>
   $ bootstrap init # install the core component and create or init gitOps repo, such as: ArgoCD, external-secret, vault, etc.
   $ bootstrap project create testing # create a new project in ArgoCD
   $ bootstrap status # check current h4 platform status, show echo component status, such as: ArgoCD, external-secret, vault, etc.
   $ bootstrap delete # delete the h4 platform core component, such as: ArgoCD, external-secret, vault, etc.
   ```
