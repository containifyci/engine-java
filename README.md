# engine-java

**engine-java** extends [engine-ci](https://github.com/containifyci/engine-ci) to provide first-class support for building Java applications using **Maven**.

This repository adds Java-specific build logic and includes an integration test project to ensure end-to-end compatibility with the engine-ci build system.

---

## Overview

engine-java enables seamless Java builds within the [ContainifyCI](https://github.com/containifyci) ecosystem.  
It leverages the same declarative build definition as engine-ci and adds Maven-specific build steps and conventions.

Key highlights:
- Maven-based build and packaging support
- Full integration with engine-ci pipelines
- Includes a complete integration test project
- Designed for extensibility and reuse across multiple Java projects

---

## How It Works

engine-java is a build extension that plugs into engine-ci’s build service.  
It detects Java projects, runs Maven commands (such as `mvn clean package`), and exports build artifacts for downstream usage or release automation.

Under the hood:
- It uses the same build execution flow and container orchestration as engine-ci.
- The build process runs inside isolated containers for reproducibility.
- Artifacts and logs are automatically collected by engine-ci.

---

## Integration Test Example

The repository includes a small Java web app used for testing:

```bash
testdata/hello-world-servlet/
````

This sample project:

* Demonstrates how a typical Java Maven project is built using engine-java.
* Is used in CI integration tests to verify that builds run successfully inside engine-ci environments.

You can manually run the integration test locally to validate your setup.

---

## ▶Running Locally

Make sure you have [engine-java](https://github.com/containifyci/engine-java) installed.

```bash
go install github.com/containifyci/engine-java@latest
```

Then you can trigger a build using:

```bash
engine-java run
```

This will automatically load engine-java and execute the Maven build steps defined in the extension.

---

## 🧱 Repository Structure

```
engine-java/
├── build/                      # Java build logic extension for engine-ci
├── testdata/
│   └── hello-world-servlet/    # Sample Maven project for integration testing
├── internal/
│   └── ...                     # Internal utilities for engine-java
└── README.md
```

---

## 🔧 Requirements

* Golang >= 1.25
* Docker or Podman (for build isolation)

---

## 🤝 Contributing

Contributions are welcome!
If you want to add new Java build features or integrations, please open a pull request or start a discussion in the [engine-ci repo](https://github.com/containifyci/engine-ci/discussions).

---

## 📄 License

Licensed under the [Apache 2.0 License](LICENSE).

---

### 🌟 Part of the ContainifyCI Ecosystem

* [engine-ci](https://github.com/containifyci/engine-ci) – Core build engine
* [Engine Go](https://github.com/containifyci/engine-go) – Go build extension
* [engine-java](https://github.com/containifyci/engine-java) – Maven/Java build extension

```

---

Would you like me to include a short “example output” section (like a sample `engine-ci` log showing Maven build steps), or keep it concise and clean like above?
```
