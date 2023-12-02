# ENVITE

ENVITE helps define and provision environments for development, testing, and continuous integration.

It's designed to allow switching between local development needs and automated environments seamlessly
and provide the best tooling to describe, provision, and monitor integrated components.

A demo screen capture can be found [here](/demo.mov).

---

## Motivation

Any software that interacts with external components requires some solutions. In production environments you need
these two components to be able to interact in a safe and secure fashion. Obviously, there are countless solutions
for these needs such as cloud-managed products, container orchestration solutions such as Kubernetes, load balancers,
service discovery solutions, service mesh products, and so on. ENVITE doesn't aim to help there.

Next, if you want to run non-production automation such as integration testing during a CI/CD pipeline, you're going
to need to create an environment similar to production to be able to execute these tests.

Lastly, development processes require a similar environment to function properly.

ENVITE aims to help with the last 2 use cases - development, and non-production automation. It aims to make them
completely similar to allow full reproducibility, on the one hand, and the best tooling for development needs, on the
other hand.

## Alternatives

#### So why not stick to what you use today?

For starters, you might want to do that. Let's see when you actually need ENVITE.
Here are the popular alternatives and how they compare with ENVITE.

##### Using Kubernetes for production, CI, and development

This method has a huge advantage: you only have to describe your environment once. This means you maintain only
one description of your environments - using Kubernetes manifest files, but more importantly, the way your components
are deployed and provisioned in production is identical to the way they are in development in CI.

Let's talk about some possible downsides:
* Local development is not always intuitive. While actively working on one or more components, there are some issues 
to solve:
  * If you fully containerize everything, like you normally do in Kubernetes:
    * You need to solve how you debug running containers. Attaching a remote debugging session is not always easy.
    * How do you rebuild container images each time you perform a code change? This process can take several minutes
every time you perform any code change.
    * How do you manage and override image tag values in your original manifest files? Does it mean you maintain
separate manifest files for production and dev purposes?
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
that prevents writing custom code or containerizing your code. Either way, you add complexity and time.
* What about an environment that keeps some components outside Kubernetes in production? For instance, some companies
do not run their databases inside a Kubernetes cluster. This also means you have to maintain manifest files
specifically for dev and CI, and the environments are not identical to production.

##### Using docker-compose for CI and development

If you're not using container orchestration tools like Kubernetes in production, but need some integration
between several components, this will probably be your first choice.

However, it does have all the possible downsides of Kubernetes mentioned above, on top of some other ones:
* You manage your docker-compose manifest files specifically for dev and CI. This means you have a duplicate to
maintain, but also, your dev and CI envs can potentially be different from production.
* Managing dependencies between services is not always easy - if one service needs to be fully operational before
another one starts, it can be a bit tricky.

##### Using a remote staging/dev environment

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

##### Using testcontainers or a similar library to write container management code for CI and development

This option is quite close to ENVITE. These tools allow you to write custom code to describe your environment,
so you have full control over what you can do. As with most other options, you must manage your test env separately from
production since you don't use testcontainers in production. This means you have to maintain 2 copies, but also,
production env can defer from your test env.

In addition, testcontainers have 2 more downsides:
* You can only write in Java or Go.
* testcontainers bring a LOT of dependencies.

##### Can ENVITE meet my needs?

ENVITE is quite close to testcontainers. It allows you to either write Go, or use YAML files to describe your env.
It can be used as a Go library, or as a CLI tool directly without actually writing code.

ENVITE is designed around running seamlessly inside and outside docker containers, to allow simple debugging of
components you currently work on, while running all the rest in containers, connecting everything fluently, and
providing you with the best tooling to manage and monitor the entire environment.

At this point, you will have to manage ENVITE files/code separately from your production environment, but we do want
to add support to read directly from Helm and Kustomize files to allow maintaining only one copy of production env.

One last thing to consider, since ENVITE runs everything locally, you can't run too many components on your machine.
As mentioned earlier, if you need 50 components up and running to run your tests, running it all locally is not
feasible. You will have to use multiple remote machines for that. If this is your use case, an interesting company
that does it well is [Raftt](https://www.raftt.io/) and I suggest reading more about what they do.

## Local Development

To locally work on the UI, cd into the `ui` dir and run react dev server using `npm start`.

To build the UI into shipped static files run `./build-ui.sh`.
