## What are Microservices ?

- Microservices is an architectural style where an application is divided into small, independent, and loosely coupled services, each responsible for a specific business capability.
- Each microservice has its own codebase, allowing a small team to develop, test, maintain, and deploy it independently.
- Microservices communicate with one another through well-defined APIs (such as REST, gRPC) or asynchronous messaging systems (such as Kafka or RabbitMQ).
- Since each service is independent, it can be deployed, updated, and scaled without affecting the rest of the application.
- Microservices are technology-agnostic, meaning each service can use the most suitable programming language, framework, or database based on its requirements.
- Following the Database per Service pattern, every microservice owns its own database, ensuring loose coupling and preventing direct database sharing between services.