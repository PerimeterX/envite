# ENVITE

<img alt="envite-logo" src="https://raw.githubusercontent.com/PerimeterX/envite/assets/logo-small.svg">

[![CodeQL Status](https://img.shields.io/github/actions/workflow/status/perimeterx/envite/codeql.yml?branch=main&logo=github&label=CodeQL)](https://github.com/PerimeterX/envite/actions/workflows/codeql.yml?query=branch%3Amain++)
[![Run Tests](https://img.shields.io/github/actions/workflow/status/perimeterx/envite/go.yml?branch=main&logo=github&label=Run%20Tests)](https://github.com/PerimeterX/envite/actions/workflows/go.yml?query=branch%3Amain)
[![Dependency Review](https://img.shields.io/github/actions/workflow/status/perimeterx/envite/dependency-review.yml?logo=github&label=Dependency%20Review)](https://github.com/PerimeterX/envite/actions/workflows/dependency-review.yml?query=branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/perimeterx/envite)](https://goreportcard.com/report/github.com/perimeterx/envite)
[![Go Reference](https://pkg.go.dev/badge/github.com/perimeterx/envite.svg)](https://pkg.go.dev/github.com/perimeterx/envite)
[![Licence](https://img.shields.io/github/license/perimeterx/envite)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/perimeterx/envite)](https://github.com/PerimeterX/envite/releases)
![Top Languages](https://img.shields.io/github/languages/top/perimeterx/envite)
[![Issues](https://img.shields.io/github/issues-closed/perimeterx/envite?color=%238250df&logo=github)](https://github.com/PerimeterX/envite/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr-closed-raw/perimeterx/envite?color=%238250df&label=merged%20pull%20requests&logo=github)](https://github.com/PerimeterX/envite/pulls)
[![Commits](https://img.shields.io/github/last-commit/perimeterx/envite)](https://github.com/PerimeterX/envite/commits/main)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

A framework to manage development and testing environments.

## Contents

* [Why ENVITE?](#why-envite)
  - [Kubernetes](#kubernetes)
  - [Docker Compose](#docker-compose)
  - [Remote Staging/Dev Environments](#remote-stagingdev-environments)
  - [TestContainers](#testcontainers)
  - [How's ENVITE Different?](#hows-envite-different)
  - [Does ENVITE Meet My Need?](#does-envite-meet-my-need)
* [Usage](#usage)
  - [Go SDK Usage](#go-sdk-usage)
  - [CLI Usage](#cli-usage)
  - [Demo](#demo)
  - [Execution Modes](#execution-modes)
  - [Flags and Options](#flags-and-options)
  - [Adding Custom Components](#adding-custom-components)
* [Key Elements of ENVITE](#key-elements-of-envite)
* [Local Development](#local-development)
* [Contact and Contribute](#contact-and-contribute)
* [ENVITE Logo](#envite-logo)

## Why ENVITE?

<img align="right" width="200" alt="marshmallow-gopher" src="https://raw.githubusercontent.com/PerimeterX/envite/assets/logo3.svg">

Why should I Choose ENVITE? Why not stick to what you have today?

For starters, you might want to do that. Let's see when you **actually** need ENVITE.
Here are the popular alternatives and how they compare with ENVITE.

#### Kubernetes

Using Kubernetes for testing and development, in addition to production.

This method has a huge advantage: you only have to describe your environment once. This means you maintain only
one description of your environments - using Kubernetes manifest files. But more importantly, the way your components
are deployed and provisioned in production is identical to the way they are in development in CI.

Let's talk about some possible downsides:
* Local development is not always intuitive. While actively working on one or more components, there are some issues
  to solve:
  * If you fully containerize everything, like you normally do in Kubernetes:
    * You need to solve how you debug running containers. Attaching a remote debugging session is not always easy.
    * How do you rebuild container images each time you perform a code change? This process can take several minutes
      every time you perform any code change.
    * How do you manage and override image tag values in your original manifest files? Does it mean you maintain
      separate manifest files for production and dev purposes? Do you have to manually override environment variables?
    * Can you provide hot reloading or similar tools in environments where this is desired?
  * If you choose to avoid developing components in containers, and simply run them outside the cluster:
    * How easy it is to configure and run a component outside the cluster?
    * Can components running outside the cluster communicate with components running inside it? This needs to be solved
      specifically for development purposes.
    * Can components running inside the cluster communicate with components running outside of it? This requires
      a different, probably more complex solution.
* What about non-containerized steps? By that, I'm not referring to actual production components that are not
  containerized. I'm talking about steps that do not exist in production at all. For instance, creating seed data
  that must exist in a database for other components to boot successfully. This step usually involves writing some code
  or using some automation to create initial data. For each such requirement, you can either find a solution
  that prevents writing custom code or containerizing your code. Either way, you add complexity and time. But more
  importantly, it misses the original goal of dev and CI being identical to production.
  These are extra steps that can hide production issues, or create issues that aren't really there.
* What about an environment that keeps some components outside Kubernetes in production? For instance, some companies
  do not run their databases inside a Kubernetes cluster. This also means you have to maintain manifest files
  specifically for dev and CI, and the environments are not identical to production.

#### Docker Compose

If you're not using container orchestration tools like Kubernetes in production, but need some integration
between several components, this will probably be your first choice.

However, it does have all the possible downsides of Kubernetes mentioned above, on top of some other ones:
* You manage your docker-compose manifest files specifically for dev and CI. This means you have a duplicate to
  maintain, but also, your dev and CI envs can potentially be different from production.
* Managing dependencies between services is not always easy - if one service needs to be fully operational before
  another one starts, it can be a bit tricky.

#### Remote Staging/Dev Environments

Some cases have developers use a remote environment for dev and testing purposes.
It can either be achieved using tools such as Kubernetes, or even simply connecting to remote components.

These solutions are a very good fit for use cases that require running a lot of components. I.e., if you need 50
components up and running to run your tests, running it all locally is not feasible.
However, they can have downsides or complexities:
* You need internet connectivity. It sounds quite funny because you have a connection everywhere these days, right?
  But think about the times that your internet goes down, and you can at least keep on debugging your if statement.
  Now you can't. Think about all the times that the speed goes down, this directly affects your ability to run and debug
  your local code.
* What if something breaks? Connecting to remote components every time you want to do any kind of local development
  simply add issues that are more complex to understand and debug. You might need to debug your debugging sessions.
* Is this environment shared? If so, this is obviously bad. Tests can suddenly stop passing because someone made
  a change that had unintended consequences.
* If this environment is not shared, how much does it cost to have an entire duplicate of the production stack for each
  engineer in the organization?

#### TestContainers

This option is quite close to ENVITE. TestContainers and similar tools allow you to write custom code to describe your
environment, so you have full control over what you can do. As with most other options, you must manage your test env
separately from production since you don't use testcontainers in production. This means you have to maintain 2 copies,
but also, production env can defer from your test env.

In addition, testcontainers have 2 more downsides:
* You can only write in Java or Go.
* testcontainers bring a LOT of dependencies.

#### How's ENVITE Different?

With either option you choose, the main friction you're about to encounter is debugging and local development. Suppose
your environment contains 10 components, but you're currently working on one. You make changes that you want to quickly
update, you debug and use breakpoints, you want hot reloading or other similar tools - either way,
if you must use containers it's going to be harder. ENVITE is designed to make it simple.

ENVITE supports a Go SDK that resembles testcontainers and a YAML CLI tool that resembles docker-compose. However,
containers are not a requirement. ENVITE is designed to allow components to run inside or outside containers.
Furthermore, ENVITE components can be anything, as long as they implement a simple interface.
Components like data seed steps do not require containerizing at all.
This allows the simple creation of components and ease of debugging and local development.
It connects everything fluently and provides the best tooling to manage and monitor the entire environment.
You'll see it in action below.

#### Does ENVITE Meet My Need?

As with other options, it does mean your ENVITE description of the environment is separate from the definition of
production environments. If you want to know what's the best option for you - If you're able to run testing and local
dev using only production manifest files, and also able to easily debug and update components,
and this solution is cost-effective - you might not need ENVITE.
If this is not the case, ENVITE is probably worth checking out.

At some point, we plan to add support to read directly from Helm and Kustomize files to allow enjoying the goodies
of ENVITE without having to maintain a duplicate of production manifests.

Another limitation of ENVITE (and most other options as well) - since it runs everything locally, there's a limit on
the number of components it can run. If you need 50 components up and running to run your tests,
running it all locally might not be feasible.

If this is your direction, another interesting alternative to check out is [Raftt](https://www.raftt.io/).

## Usage

ENVITE offers flexibility in environment management through both a Go SDK and a CLI.
Depending on your use case, you can choose the method that best fits your needs.

The Go SDK provides fine-grained control over environment configuration and components.
For example, you can create conditions to determine what the environment looks like or create a special connection
between assets, particularly in seed data.
However, the Go SDK is exclusively applicable within a Go environment and is most suitable for organizations
or individuals already using or open to incorporating it into their tech stack.
Regardless of the programming languages employed, if you opt to write your tests in Go,
the ENVITE Go SDK is likely a more powerful choice.

Otherwise, the CLI is an easy-to-install and intuitive alternative, independent of any tech stack,
and resembles docker-compose in its setup and usage. However, it's more powerful than docker-compose in many use cases as [mentioned above](#docker-compose).

#### Go SDK Usage

```go
package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/perimeterx/envite"
	"github.com/perimeterx/envite/docker"
	"github.com/perimeterx/envite/seed/mongo"
)

func runTestEnv() error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	network, err := docker.NewNetwork(dockerClient, "docker-network-or-empty-to-create-one", "my-test-env")
	if err != nil {
		return err
	}

	persistence, err := network.NewComponent(docker.Config{
		Name:    "mongo",
		Image:   "mongo:7.0.5",
		Ports:   []docker.Port{{Port: "27017"}},
		Waiters: []docker.Waiter{docker.WaitForLog("Waiting for connections")},
	})
	if err != nil {
		return err
	}

	cache, err := network.NewComponent(docker.Config{
		Name:    "redis",
		Image:   "redis:7.2.4",
		Ports:   []docker.Port{{Port: "6379"}},
		Waiters: []docker.Waiter{docker.WaitForLog("Ready to accept connections tcp")},
	})
	if err != nil {
		return err
	}

	seed := mongo.NewSeedComponent(mongo.SeedConfig{
		URI: fmt.Sprintf("mongodb://%s:27017", persistence.Host()),
	})

	env, err := envite.NewEnvironment(
		"my-test-env",
		envite.NewComponentGraph().
			AddLayer(map[string]envite.Component{
				"persistence": persistence,
				"cache":       cache,
			}).
			AddLayer(map[string]envite.Component{
				"seed": seed,
			}),
	)
	if err != nil {
		return err
	}

	server := envite.NewServer("4005", env)
	return envite.Execute(server, envite.ExecutionModeDaemon)
}
```

#### CLI Usage

1. Install ENVITE from the [GitHub releases page](https://github.com/PerimeterX/envite/releases/latest).
2. Create an envite.yml file:
```yaml
default_id: "my-test-env"
components:
  -
    persistence:
      type: docker component
      image: mongo:7.0.5
      name: mongo
      ports:
        - port: '27017'
      waiters:
        - string: Waiting for connections
          type: string
    cache:
      type: docker component
      image: redis:7.2.4
      name: redis
      ports:
        - port: '6379'
      waiters:
        - string: Ready to accept connections tcp
          type: string
  -
    seed:
      type: mongo seed
      uri: mongodb://{{ persistence }}:27017
      data:
        - db: data
          collection: users
          documents:
            - first_name: John
              last_name: Doe
```
3. Run ENVITE: `envite`.

The full list of CLI supported components can be found [here](https://github.com/PerimeterX/envite/blob/b069952815519b3026551485af9e63be1bdca751/cmd/envite/environment.go#L68).

#### Demo

With either approach, the result is a UI served via the browser. It enables managing the environment, monitoring,
initiating and halting components, conducting detailed inspections, debugging, and providing all essential tools
for development and testing, as well as automated and CI/CD processes.

[![ENVITE Demo](https://raw.githubusercontent.com/PerimeterX/envite/assets/demo.gif)](https://raw.githubusercontent.com/PerimeterX/envite/assets/demo.mp4)

Voilà! You now have a fully usable dev and testing environment.

#### Execution Modes

ENVITE supports three execution modes:

* Daemon Mode (`envite -mode start`): Start execution mode, which starts all components in the environment,
and then exits.
* Start Mode (`envite -mode stop`): Stops all components in the environment, performs cleanup, and then exits.
* Stop Mode (`envite -mode daemon`): Starts ENVITE as a daemon and serves a web UI.

Typically, the `daemon` mode will be used for local purposes, and a combination of `start` and `stop` modes will be
used for Continuous Integration or other automated systems.

#### Flags and Options

All flags and options are described via envite -help command:

```bash
  mode
        Mode to operate in (default: daemon)
  -file value
        Path to an environment yaml file (default: `envite.yml`)
  -id value
        Override the environment ID provided by the environment yaml
  -network value
        Docker network identifier to be used. Used only if docker components exist in the environment file. If not provided, ENVITE will create a dedicated open docker network.
  -port value
        Web UI port to be used if mode is daemon (default: `4005`)
```

#### Adding Custom Components

Integrate your own components into the environment, either as Docker containers or by providing implementations
of the [envite.Component](https://github.com/PerimeterX/envite/blob/b4e9f545226c990a1025b9ca198856faff8b5eed/component.go#L13) interface.

## Key Elements of ENVITE

ENVITE contains several different elements:
* `Environment`: Represents the entire configuration, containing components and controlling them to provide a fully
functional environment.
* `Component`: Represents a specific part of the environment, such as a Docker container or a custom component.
* `Component` Graph: Organizes components into layers and defines their relationships.
* `Server`: Allow serving a UI to manage the environment.

## Local Development

To locally work on ENVITE UI, cd into the `ui` dir and run react dev server using `npm start`.

To build the UI into shipped static files run `./build-ui.sh`.

## Contact and Contribute

Reporting issues and requesting features may be done on our [GitHub issues page](https://github.com/PerimeterX/envite/issues).
For any further questions or comments, you can reach us at [open-source@humansecurity.com](mailto:open-source@humansecurity.com).

Any type of contribution is warmly welcome and appreciated ❤️
Please read our [contribution](CONTRIBUTING.md) guide for more info.

If you're looking for something to get started with, you can always follow our [issues page](https://github.com/PerimeterX/envite/issues) and look for
[good first issue](https://github.com/PerimeterX/envite/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) and
[help wanted](https://github.com/PerimeterX/envite/issues?q=is%3Aissue+label%3A%22help+wanted%22+is%3Aopen) labels.

## ENVITE Logo

ENVITE logo and assets by [Adva Rom](https://www.linkedin.com/in/adva-rom-7a6738127/) are licensed under a <a rel="license" href="http://creativecommons.org/licenses/by/4.0/">Creative Commons Attribution 4.0 International License</a>.<br />
