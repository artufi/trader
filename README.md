# 🧠 Real-Time Trading Bot for XTB WebSocket API

⚠️ **Note**: The XTB API is no longer available. This project has been open-sourced to demonstrate the architecture, 
design principles, and development approach behind a real-time trading bot.

---

> **TL;DR**: A real-time trading bot written in Go, built specifically for the XTB WebSocket API.  
> Supports multiple users, concurrent trading clients, and AI-generated signals.  
> Designed with clean code principles, high responsiveness, and future scalability in mind.  
> Now released as an educational project and development reference.

---

## 🚀 About the Project

This bot was built to connect to the XTB trading platform using WebSocket for real-time market data and trade execution.  
Each user could subscribe to specific strategies or timeframes, and the bot would respond to signals accordingly  
by spawning lightweight, fast-response clients to execute trades independently and in parallel, 
ensuring low latency and high reliability.

The internal API was implemented in REST for simplicity and early testing. In the future, 
**streaming or event-driven communication** could replace REST to lower signal latency.

The broader vision included a move to the cloud, infrastructure automation, CI/CD pipelines and full observability.

---

## ⚙️ Core Features

- ✅ Trading across multiple timeframes.
- ✅ Parallel execution of lightweight trading clients.
- ✅ Real-time communication with the XTB API over WebSocket protocol.
- ✅ Routing of AI-generated signals (ML model) to matching strategy subscribers.
- ✅ Internal tracking of trades and cleanup logic.
- ✅ Straightforward, extendable code structure.

---

## 💭 Why Open Source?

This project was developed over time as a personal and experimental system. Once XTB discontinued access to its API,
I decided to open source the codebase - both to share how such a bot can be built, and to provide a reference
for backend design in Go.

While some parts are incomplete or require refinement, the overall structure shows how to approach real-time systems,
signal handling, and backend design.

The goal is for this project to serve as a practical, understandable foundation for others to explore and learn from.

---

## 🔮 Future Plans

The original goal was to eventually build a platform where users could subscribe to strategies or timeframes, 
and the bot would act on their behalf in real time.

Planned or potential future improvements include:

- Support for multiple trading APIs to avoid dependency on a single provider.
- Retry logic and robust error handling - critical when real funds are involved.
- Transition from REST to event-driven or streaming-based communication for lower latency.
- Full observability (metrics, logs, traces) using OpenTelemetry.
- CI/CD pipelines for testing and deployment.
- Adding Kubernetes, Helm, and Terraform.
- Additional test coverage - unit and integration - is still being added.

---

## 🛠️ Tech Stack

- **Go (Golang)** - backend logic and concurrency.
- **PostgreSQL** - persistent storage of trades and user sessions.
- **Goose** - database schema migrations.
- **Dotenv** - for loading environment variables from `.env` file during development.
- **Docker & Docker Compose** - for local development and deployment.
- **WebSocket (XTB)** - used for market data and trading interface.
- **REST** - for simple signal testing and user operations.
- **OpenTelemetry** (planned) - for logs, metrics, and traces.
- **Kubernetes + Helm** (planned) - for orchestration and deployment.
- **Terraform** (planned) - infrastructure as code.
- **CI/CD** (planned) - to automate testing and deployment pipelines.

---

## 🚧 Current Limitations

This project is still a prototype. Known limitations include:

- One hardcoded test user.
- No user management or authentication services.
- Retry logic and recovery from edge cases are not yet implemented.
- Missing CI/CD setup and full observability.
- Not yet production-ready - especially for handling real capital without additional safety layers.
- Test coverage is partial, and is being expanded as the codebase evolves.
- Requires optimization of database schema (e.g., data types, indexes) for long-term scalability and performance.

---

## 🤝 Want to Collaborate?

I'm open to feedback, code reviews, or collaborating with others - whether it's about improving the project, 
sharing ideas, or building something new together.
Feel free to reach out!
