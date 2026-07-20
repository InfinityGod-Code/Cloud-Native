## What are Containers ?

1. Traditional monolithic applications are packaged and deployed as a single unit. Even a small change requires redeploying the entire application, which can lead to downtime, difficult rollbacks, and limited scalability.
2. Microservices architecture addresses these challenges by breaking an application into smaller, independent services. Containers provide the ideal way to package and run these services consistently across environments.
3. A container is a lightweight, portable package that includes an application along with everything it needs to run, such as its runtime, libraries, dependencies, and configuration.
4. Containers ensure that an application behaves the same way in development, testing, and production, eliminating the common "it works on my machine" problem.
5. Container images can be built once and deployed anywhere that supports a container runtime (such as Docker), making deployments fast, reliable, and reproducible.
6. In a microservices architecture, each service is typically packaged in its own container. This enables:
- Independent development and testing
- Independent deployment and updates
- Independent scaling based on demand
- Better fault isolation
7. Containers communicate with each other through well-defined APIs or network interfaces, allowing multiple services to work together while remaining loosely coupled.