---
title: "Design Principles: Systems"
description: The principles and patterns for composing containers into a
  system, each as the question to ask of its design
slug: guide-design-principles-systems
type: guide
status: active
---

# Design principles: systems

Ask these questions when you design how a system's containers integrate, or
review that design. Use them with the universal guide,
[[guide-design-principles]], which explains how to choose and name patterns
and notes that these lists are incomplete.

| Entry | Progenitor | Questions it answers |
|-------|-------------|----------------------|
| Conway's Law | Conway, "How Do Committees Invent?", *Datamation* (1968) | Do container boundaries match the teams that own them? |
| Database per Service | Richardson, *Microservices Patterns* (2018), Database per Service | Which container owns each store? Does any other container read it directly? |
| Bounded Context and Context Map | Evans, *Domain-Driven Design* (2003), Bounded Context and Context Map | Where does one model's language end, and how do neighboring models relate? |
| Anti-Corruption Layer | Evans, *Domain-Driven Design* (2003), Anticorruption Layer | How does a container use a foreign or legacy model without adopting it? |
| Event-Driven Architecture, Publish-Subscribe | Hohpe and Woolf, *Enterprise Integration Patterns* (2003), Publish-Subscribe Channel | Which interactions should be asynchronous, so a new consumer joins without the producer changing? |
| API Gateway | Richardson, *Microservices Patterns* (2018), API Gateway | How do clients reach many containers through one entry point? |
| Backend for Frontend | Calçado, "The Back-end for Front-end Pattern (BFF)" (2015) | Does each kind of client get an entry point shaped for it? |
| Sidecar | Burns and Oppenheimer, "Design Patterns for Container-based Distributed Systems" (2016) | How does a runtime concern attach to a container without changing its code? |
| Saga | Garcia-Molina and Salem, "Sagas" (1987) | How does a business transaction that spans containers stay consistent without a distributed transaction? |
| Transactional Outbox | Richardson, *Microservices Patterns* (2018), Transactional Outbox | How does a container change its state and publish the event atomically? |
| Idempotent Consumer | Hohpe and Woolf, *Enterprise Integration Patterns* (2003), Idempotent Receiver | What happens when a message arrives twice? |
| CQRS | Young, "CQRS Documents" (2010) | Do reads and writes need separate models or stores? |
| Circuit Breaker and Bulkhead | Nygard, *Release It!* (2007), Circuit Breaker and Bulkheads | How does one container's failure stay contained? |
| Timeouts and Retry with Backoff | Nygard, *Release It!* (2007), Timeouts | How long does a caller wait, and how does it retry without a storm? |

## References

- [[guide-design-principles]] — the universal principles, and how to choose
  and name patterns
