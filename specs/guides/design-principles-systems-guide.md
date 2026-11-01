---
title: "Design Principles: Systems"
description: The principles and patterns for composing containers into a
  system, each as the question to ask of a System Blueprint
slug: guide-design-principles-systems
type: guide
status: active
---

# Design principles: systems

Ask these questions when you design how a system's containers integrate, or
audit its System Blueprint. Use them with the universal guide,
[[guide-design-principles]], which also explains how to choose patterns and
notes that these lists are incomplete. Record each pattern the design adopts
under
[[blueprint-content-standards#Name each pattern with its progenitor, constraint, and need]].

| Entry | Progenitors | Questions it answers |
|-------|-------------|----------------------|
| Conway's Law | Conway, *Datamation* (1968); Skelton and Pais, *Team Topologies* (2019) | Do container boundaries match the teams that own them? |
| Database per Service | Richardson, *Microservices Patterns* (2018) | Which container owns each store? Does any other container read it directly? |
| Bounded Context and Context Map | Evans, *Domain-Driven Design* (2003) | Where does one model's language end, and how do neighboring models relate? |
| Anti-Corruption Layer | Evans (2003) | How does a container use a foreign or legacy model without adopting it? |
| Event-Driven Architecture, Publish-Subscribe | Hohpe and Woolf, *Enterprise Integration Patterns* (2003) | Which interactions should be asynchronous, so a new consumer joins without the producer changing? |
| API Gateway and Backend for Frontend | Richardson (2018); Calçado, SoundCloud (2015) | How do clients reach many containers through one entry point shaped for each client? |
| Sidecar | Burns and Oppenheimer (2016) | How does a runtime concern attach to a container without changing its code? |
| Saga | Garcia-Molina and Salem (1987); Richardson (2018) | How does a business transaction that spans containers stay consistent without a distributed transaction? |
| Transactional Outbox | Richardson (2018) | How does a container change its state and publish the event atomically? |
| Idempotent Consumer | Hohpe and Woolf, as Idempotent Receiver (2003) | What happens when a message arrives twice? |
| CQRS | Young (2010), from Meyer's Command-Query Separation | Do reads and writes need separate models or stores? |
| Circuit Breaker and Bulkhead | Nygard, *Release It!* (2007) | How does one container's failure stay contained? |
| Timeouts and Retry with Backoff | Nygard (2007); Metcalfe and Boggs, Ethernet backoff (1976) | How long does a caller wait, and how does it retry without a storm? |

## References

- [[guide-design-principles]] — the universal guide, and how to choose patterns
- [[blueprint-content-standards]] — how a blueprint names patterns, and the
  blueprint audit
